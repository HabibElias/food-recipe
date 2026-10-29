#!/usr/bin/env python3
"""Validate the full deployment using disposable LOCAL Docker resources only.

Run with Go 1.26 and Node 22+ on PATH:
  python3 deploy/hasura/smoke_test.py
GO_BINARY can select the Go executable. Requires the managed local Docker socket.
No production URLs/credentials are consumed; all test identities are ephemeral.
"""
import contextlib
import io
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import socket
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from unittest import mock

from bootstrap import DeploymentError, Hasura, ROOT, digest, initialize

FRONTEND = ROOT.parent / "food-recipe-app"
DOCKER = ["docker", "--host=unix:///var/run/docker.sock"]


def command(args, *, data=None, env=None):
    result = subprocess.run(args, input=data, env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if result.returncode:
        # Docker arguments and provider errors may contain local test credentials.
        raise RuntimeError("Local validation command failed; command output was suppressed")
    return result.stdout.strip()


def free_port():
    with socket.socket() as handle:
        handle.bind(("0.0.0.0", 0))
        return handle.getsockname()[1]


def wait_until(check, description):
    for _ in range(60):
        try:
            if check():
                return
        except (OSError, DeploymentError):
            pass
        time.sleep(0.5)
    raise RuntimeError("Timed out waiting for " + description)


def graphql(base, query, variables=None, token=None, admin=None, denied=False):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if admin:
        headers["x-hasura-admin-secret"] = admin
    request = urllib.request.Request(base + "/v1/graphql", data=json.dumps({"query": query, "variables": variables or {}}).encode(), headers=headers)
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            body = json.load(response)
    except urllib.error.HTTPError as error:
        body = json.load(error)
    if denied:
        assert body.get("errors"), "Expected GraphQL access denial"
    else:
        assert not body.get("errors"), "GraphQL operation failed; response details suppressed"
    return body.get("data")


def run_checks(base, admin):
    def gql(query, variables=None, token=None, privileged=False, denied=False):
        return graphql(base, query, variables, token, admin if privileged else None, denied)

    def account(tag):
        name = "smoke" + tag + secrets.token_hex(4)
        email, password = name + "@example.invalid", secrets.token_urlsafe(18)
        result = gql("mutation($input: SignupInput!) { signup(input:$input) { token createdUser {id email username FirstName bio created_at} } }", {"input": {"email": email, "password": password, "username": name, "first_name": "Smoke", "last_name": "User"}})["signup"]
        profile = result["createdUser"]
        assert profile["email"] == email
        logged = gql("mutation($input:LoginInput!){login(input:$input){token loggedInUser{id email}}}", {"input": {"email": email, "password": password}})["login"]
        assert logged["loggedInUser"]["id"] == profile["id"]
        me = gql("mutation($input:MeInput!){Me(input:$input){token user{id email bio created_at}}}", {"input": {"token": logged["token"]}})["Me"]
        assert me["user"]["id"] == profile["id"] and me["user"]["email"] == email
        return profile["id"], logged["token"]

    alice, ta = account("a")
    bob, tb = account("b")
    print("signup/login/Me actions and authenticated profile outputs: passed")
    for column in ["email", "password"]:
        gql("{user{" + column + "}}", denied=True)
        gql("{user{" + column + "}}", token=ta, denied=True)
    print("public profile email/password exclusion: passed")

    changed = gql('mutation($id:uuid!){update_user(where:{id:{_eq:$id}},_set:{bio:"A little from my kitchen"}){affected_rows}}', {"id": alice}, token=ta)
    assert changed["update_user"]["affected_rows"] == 1
    denied = gql('mutation($id:uuid!){update_user(where:{id:{_eq:$id}},_set:{bio:"Not your kitchen"}){affected_rows}}', {"id": alice}, token=tb)
    assert denied["update_user"]["affected_rows"] == 0
    gql('mutation{update_user(where:{},_set:{email:"wrong@example.invalid"}){affected_rows}}', token=ta, denied=True)
    print("own-profile persistence and cross-user/private-field update denial: passed")

    defaults = {"limit": 10, "offset": 0, "where": {}, "id": -1, "user_id": alice}
    count = 0
    for file in sorted((FRONTEND / "server/_queries").glob("*.gql")):
        if file.name == "GetMe.gql":
            continue
        document = file.read_text()
        variables = {name: defaults[name] for name in re.findall(r"\$(\w+)\s*:", document.split("{", 1)[0]) if "$" + name in document.split("{", 1)[1]}
        gql(document, variables, token=None if "NotLoggedIn" in file.name else ta)
        count += 1
    assert count == 9, "Expected all nine frontend recipe documents"
    print("all nine frontend recipe query documents: passed")

    category = gql("{category(limit:1){id}}", privileged=True)["category"][0]["id"]
    create = (FRONTEND / "server/_mutations/InsertRecipeMutation.gql").read_text()
    values = {"title": "Smoke recipe", "category_id": category, "description": "Test", "is_paid": False, "prep_time": 10, "price": 12.34, "recipe_ingredients": {"data": [{"ingredient": "private ingredient"}]}, "recipe_steps": {"data": [{"step_no": 1, "step_description": "private step"}]}, "user_id": alice}
    free = gql(create, values, token=ta)["insert_recipe"]["returning"][0]["id"]
    paid = gql(create, {**values, "is_paid": True}, token=ta)["insert_recipe"]["returning"][0]["id"]
    gql(create, values, token=tb, denied=True)
    gql(create, values, denied=True)
    for operation in ["update_recipe(where:{id:{_eq:" + str(free) + '}},_set:{title:"hijack"})', "delete_recipe(where:{id:{_eq:" + str(free) + "}})"]:
        assert gql("mutation{" + operation + "{affected_rows}}", token=tb)[operation.split("(")[0]]["affected_rows"] == 0
    gql('mutation{insert_recipe_image(objects:{recipe_id:' + str(free) + ',img_url:"https://example.invalid/hijack",is_thumbnail:false}){affected_rows}}', token=tb, denied=True)
    print("frontend nested recipe insertion and cross-user write denial: passed")

    content = "query($id:Int!){recipe(where:{id:{_eq:$id}}){recipe_steps{step_description} recipe_ingredients{ingredient}}}"
    for token in [None, tb]:
        visible = gql(content, {"id": free}, token=token)["recipe"][0]
        assert visible["recipe_steps"] and visible["recipe_ingredients"]
        hidden = gql(content, {"id": paid}, token=token)["recipe"][0]
        assert not hidden["recipe_steps"] and not hidden["recipe_ingredients"]
    assert gql(content, {"id": paid}, token=ta)["recipe"][0]["recipe_steps"]
    gql('mutation($uid:uuid!,$id:Int!){insert_payment(objects:{user_id:$uid,recipe_id:$id,tx_ref:"smoke",amount:10,currency:"ETB",chapa_checkout_url:"https://example.invalid",status:"success"}){affected_rows}}', {"uid": bob, "id": paid}, privileged=True)
    assert gql(content, {"id": paid}, token=tb)["recipe"][0]["recipe_steps"]
    print("paid content restricted to owner/successful buyer: passed")

    gql("mutation($uid:uuid!,$id:Int!){insert_bookmark_one(object:{user_id:$uid,recipe_id:$id}){id}}", {"uid": alice, "id": free}, token=ta)
    assert gql("{bookmark{id}}", token=ta)["bookmark"]
    assert not gql("{bookmark{id}}", token=tb)["bookmark"]
    gql("{bookmark{id}}", denied=True)
    gql('mutation($uid:uuid!,$id:Int!){insert_recipe_comment(objects:{user_id:$uid,recipe_id:$id,comment_body:"Looks great"}){affected_rows} insert_recipe_rating(objects:{user_id:$uid,recipe_id:$id,rating:5}){affected_rows}}', {"uid": bob, "id": free}, token=tb)
    gql("mutation($uid:uuid!,$id:Int!){insert_recipe_like(objects:{user_id:$uid,recipe_id:$id}){affected_rows}}", {"uid": bob, "id": free}, token=tb)
    notifications = []
    for _ in range(30):
        notifications = gql("{notifications{id message read}}", token=ta)["notifications"]
        if notifications:
            break
        time.sleep(0.5)
    assert notifications, "Like event did not deliver notification"
    assert not gql("{notifications{id message read}}", token=tb)["notifications"]
    assert gql("mutation{update_notifications(where:{read:{_eq:false}},_set:{read:true}){affected_rows}}", token=tb)["update_notifications"]["affected_rows"] == 0
    assert gql("mutation{update_notifications(where:{read:{_eq:false}},_set:{read:true}){affected_rows}}", token=ta)["update_notifications"]["affected_rows"] > 0
    print("comments/ratings/bookmarks, like event and private notification permissions: passed")
    fields = gql("{__schema{mutationType{fields{name}}}}", token=ta)["__schema"]["mutationType"]["fields"]
    assert not {"uploadRecipeImages", "uploadUserImage", "initializePayment"} & {field["name"] for field in fields}
    print("optional upload/payment user actions: explicitly disabled")
    payload = json.dumps({"endpoint": base.replace("http://", "ws://") + "/v1/graphql", "alice": ta, "bob": tb, "notification": notifications[0]["id"]})
    print(command(["node", str(Path(__file__).with_name("subscription_smoke.mjs"))], data=payload))


def validate():
    if not FRONTEND.is_dir():
        raise RuntimeError("Sibling food-recipe-app checkout is required")
    suffix = secrets.token_hex(5)
    network = "recipe-smoke-" + suffix
    postgres, hasura = network + "-pg", network + "-hasura"
    password, admin, signing_key, action_secret = [secrets.token_hex(24) for _ in range(4)]
    api_port = free_port()
    api_process = None
    prior = {key: os.environ.get(key) for key in ["HASURA_URL", "HASURA_GRAPHQL_ADMIN_SECRET"]}
    try:
        with tempfile.TemporaryDirectory(prefix="recipe-deployment-") as temporary:
            executable = Path(temporary) / "api"
            go = os.environ.get("GO_BINARY", "go")
            env = {**os.environ, "GOTOOLCHAIN": "local", "GOMAXPROCS": "4"}
            command([go, "build", "-mod=vendor", "-o", str(executable), "./cmd/app"], env=env)
            command(DOCKER + ["network", "create", network])
            command(DOCKER + ["run", "-d", "--name", postgres, "--network", network, "--network-alias", "postgres", "-e", "POSTGRES_PASSWORD=" + password, "-e", "POSTGRES_DB=recipe_test", "postgres:15-alpine"])
            wait_until(lambda: subprocess.run(DOCKER + ["exec", postgres, "pg_isready", "-U", "postgres", "-d", "recipe_test"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0, "PostgreSQL")
            command(DOCKER + ["run", "-d", "--name", hasura, "--network", network, "--add-host", "host.docker.internal:host-gateway", "-p", "127.0.0.1::8080", "-e", "HASURA_GRAPHQL_DATABASE_URL=postgres://postgres:" + password + "@postgres:5432/recipe_test", "-e", "HASURA_GRAPHQL_ADMIN_SECRET=" + admin, "-e", "HASURA_GRAPHQL_JWT_SECRET=" + json.dumps({"type": "HS256", "key": signing_key}), "-e", "HASURA_GRAPHQL_UNAUTHORIZED_ROLE=anon", "-e", "HASURA_GRAPHQL_ENABLE_CONSOLE=false", "-e", "HASURA_GRAPHQL_DEV_MODE=false", "-e", "API_BASE_URL=http://host.docker.internal:" + str(api_port), "-e", "HASURA_ACTION_SECRET=" + action_secret, "-e", "NO_PROXY=host.docker.internal,postgres,localhost,127.0.0.1", "-e", "no_proxy=host.docker.internal,postgres,localhost,127.0.0.1", "hasura/graphql-engine:v2.44.0"])
            port = command(DOCKER + ["port", hasura, "8080/tcp"]).rsplit(":", 1)[1]
            base = "http://127.0.0.1:" + port
            os.environ.update(HASURA_URL=base + "/v1/graphql", HASURA_GRAPHQL_ADMIN_SECRET=admin)
            client = Hasura()
            wait_until(lambda: bool(client.metadata()), "Hasura")
            original_read = Path.read_text

            def invalid_migration(path, *args, **kwargs):
                text = original_read(path, *args, **kwargs)
                return text.replace("COMMIT;", "SELECT nonexistent_table_for_rollback_check;\nCOMMIT;") if path.name == "schema.sql" else text

            with mock.patch.object(Path, "read_text", invalid_migration), contextlib.redirect_stdout(io.StringIO()):
                try:
                    initialize(True)
                    raise AssertionError("Intentionally invalid migration succeeded")
                except DeploymentError:
                    pass
            rows = client.sql("SELECT table_name FROM information_schema.tables WHERE table_schema='public'", True)["result"][1:]
            assert not rows, "Failed migration left a partially initialized schema"
            with contextlib.redirect_stdout(io.StringIO()):
                initialize(True)
                initialize(True)
            print("empty database bootstrap, failed-DDL rollback and repeatability: passed")
            client.call("/v1/metadata", {"type": "create_query_collection", "args": {"name": "unrelated-object", "definition": {"queries": [{"name": "ping", "query": "query Ping { __typename }"}]}}})
            before = digest(client.metadata())
            try:
                initialize(True)
                raise AssertionError("Unknown metadata was overwritten")
            except DeploymentError:
                assert digest(client.metadata()) == before
            client.call("/v1/metadata", {"type": "drop_query_collection", "args": {"collection": "unrelated-object", "cascade": False}})
            print("unrecognized metadata rejected without mutation: passed")
            api_env = {**os.environ, "PORT": str(api_port), "HASURA_JWT_SECRET": signing_key, "HASURA_ACTION_SECRET": action_secret}
            api_process = subprocess.Popen([str(executable)], env=api_env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            wait_until(lambda: urllib.request.urlopen("http://127.0.0.1:" + str(api_port) + "/v1/healthz", timeout=2).status == 200, "API")
            run_checks(base, admin)
    finally:
        if api_process is not None:
            api_process.terminate()
            try:
                api_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                api_process.kill()
                api_process.wait()
        for container in [hasura, postgres]:
            subprocess.run(DOCKER + ["rm", "-fv", container], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        subprocess.run(DOCKER + ["network", "rm", network], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        for key, value in prior.items():
            if value is None:
                os.environ.pop(key, None)
            else:
                os.environ[key] = value


if __name__ == "__main__":
    os.chdir(ROOT)
    try:
        validate()
    except Exception as error:
        # Never dump API responses, command arguments, or ephemeral tokens.
        print("Deployment smoke test failed: " + str(error), file=sys.stderr)
        raise SystemExit(1)
