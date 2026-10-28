#!/usr/bin/env python3
"""Initialize an explicitly empty database and install this app's Hasura metadata.

Uses HASURA_URL (ending /v1/graphql) and HASURA_GRAPHQL_ADMIN_SECRET from the
process environment. No database connection string is read or printed locally.
Hasura itself resolves HASURA_GRAPHQL_DATABASE_URL, API_BASE_URL and
HASURA_ACTION_SECRET. Unknown existing schema/metadata is rejected unchanged.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import sys
import urllib.error
import urllib.parse
import urllib.request

from metadata import build_metadata

MARKER = "_food_recipe_deployment"
VERSION = 1
ROOT = Path(__file__).resolve().parents[2]


class DeploymentError(Exception):
    pass


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


class NoRedirects(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, request, fp, code, msg, headers, newurl):
        return None


class Hasura:
    def __init__(self):
        endpoint = os.environ.get("HASURA_URL", "").strip()
        parsed = urllib.parse.urlsplit(endpoint)
        if not parsed.hostname or parsed.username or parsed.password or parsed.query or parsed.fragment:
            raise DeploymentError("HASURA_URL must be a GraphQL endpoint without credentials or query parameters")
        if parsed.scheme != "https" and not (parsed.scheme == "http" and parsed.hostname in ("localhost", "127.0.0.1")):
            raise DeploymentError("HASURA_URL requires HTTPS except for local validation")
        if not parsed.path.endswith("/v1/graphql"):
            raise DeploymentError("HASURA_URL must end in /v1/graphql")
        self.base = endpoint[:-len("/v1/graphql")]
        self.opener = urllib.request.build_opener(NoRedirects())
        self.secret = os.environ.get("HASURA_GRAPHQL_ADMIN_SECRET")
        if not self.secret:
            raise DeploymentError("HASURA_GRAPHQL_ADMIN_SECRET is missing")

    def call(self, path, body):
        request = urllib.request.Request(self.base + path, data=json.dumps(body).encode(), headers={
            "Content-Type": "application/json", "x-hasura-admin-secret": self.secret,
        })
        try:
            with self.opener.open(request, timeout=60) as response:
                return json.load(response)
        except urllib.error.HTTPError as error:
            # API response/error bodies can contain submitted SQL or credential-bearing URLs.
            raise DeploymentError("Hasura operation failed with HTTP " + str(error.code)) from None
        except (urllib.error.URLError, TimeoutError, ValueError):
            raise DeploymentError("Hasura operation failed; check endpoint reachability and server logs privately") from None

    def metadata(self):
        return self.call("/v1/metadata", {"type": "export_metadata", "args": {}})

    def sql(self, sql, read_only=False):
        return self.call("/v2/query", {"type": "run_sql", "args": {
            "source": "default", "sql": sql, "read_only": read_only,
        }})


def empty_metadata(metadata):
    sources = metadata.get("sources", [])
    if len(sources) != 1 or sources[0].get("name") != "default" or sources[0].get("kind") != "postgres":
        return False
    if sources[0].get("tables") or sources[0].get("functions"):
        return False
    for key, value in metadata.items():
        if key not in ("version", "sources") and value not in (None, [], {}, {"input_objects": [], "objects": [], "scalars": [], "enums": []}):
            return False
    return True


def initialize(allow_empty):
    client = Hasura()
    current = client.metadata()
    rows = client.sql("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name", True)["result"][1:]
    tables = {row[0] for row in rows}
    if MARKER in tables:
        marker = client.sql("SELECT version::text, metadata_digest FROM public." + MARKER + " WHERE singleton = TRUE", True)["result"][1:]
        if len(marker) != 1 or marker[0][0] != str(VERSION):
            raise DeploymentError("Existing deployment marker version is unsupported; no changes applied")
        if marker[0][1] and marker[0][1] != digest(current):
            raise DeploymentError("Hasura metadata changed outside this initializer; no changes applied")
        if not marker[0][1] and not empty_metadata(current):
            raise DeploymentError("Unrecognized partial metadata initialization; no changes applied")
        print("Existing app deployment recognized; checking metadata repeatability.")
    else:
        if not allow_empty:
            raise DeploymentError("Empty-database initialization requires --initialize-empty-database")
        if tables or not empty_metadata(current):
            raise DeploymentError("Database or Hasura metadata is not empty; refusing to overwrite existing application state")
        # Both scripts are idempotent. The marker records ownership of this deployment,
        # not credentials; failed SQL is surfaced without dumping server error text.
        additions = (ROOT / "deploy/database/schema.sql").read_text().splitlines()
        statements = [i for i, line in enumerate(additions) if line.strip() and not line.lstrip().startswith("--")]
        if not statements or additions[statements[0]].strip() != "BEGIN;" or additions[statements[-1]].strip() != "COMMIT;":
            raise DeploymentError("Schema migration must have one outer BEGIN/COMMIT transaction")
        additions[statements[0]] = additions[statements[-1]] = ""
        sql = "BEGIN;\n" + (ROOT / "init.sql").read_text() + "\n" + "\n".join(additions)
        sql += "\nCREATE TABLE public." + MARKER + " (singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK(singleton), version INTEGER NOT NULL, metadata_digest TEXT); INSERT INTO public." + MARKER + " (singleton,version) VALUES (TRUE," + str(VERSION) + ");\nCOMMIT;"
        client.sql(sql)
        print("Initialized the empty PostgreSQL application schema.")
    client.call("/v1/metadata", {"type": "replace_metadata", "args": {
        "allow_inconsistent_metadata": False, "metadata": build_metadata(),
    }})
    consistency = client.call("/v1/metadata", {"type": "get_inconsistent_metadata", "args": {}})
    if not consistency.get("is_consistent"):
        raise DeploymentError("Hasura metadata is inconsistent; inspect server metadata privately")
    installed = digest(client.metadata())
    client.sql("UPDATE public." + MARKER + " SET metadata_digest = '" + installed + "' WHERE singleton = TRUE")
    print("Installed consistent Hasura metadata; optional upload/payment actions remain admin-only.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--initialize-empty-database", action="store_true")
    arguments = parser.parse_args()
    try:
        initialize(arguments.initialize_empty_database)
    except DeploymentError as error:
        print("Deployment blocked: " + str(error), file=sys.stderr)
        raise SystemExit(1)
