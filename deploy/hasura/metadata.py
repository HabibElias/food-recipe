#!/usr/bin/env python3
"""Generate versioned Hasura v2 metadata without embedding deployment secrets.

API_BASE_URL and HASURA_ACTION_SECRET are resolved by Hasura's environment.
Upload/payment actions are admin-only until their integrations and ownership
checks are enabled in a separate change.
"""
import argparse
import json
from table_metadata import build_tables


def _fields(**fields):
    return [{"name": name, "type": gql_type} for name, gql_type in fields.items()]


def _input(name, **fields):
    return {"name": name, "fields": _fields(**fields)}


def build_metadata():
    tables = build_tables()
    headers = [{"name": "X-Hasura-Action-Secret", "value_from_env": "HASURA_ACTION_SECRET"}]
    next(t for t in tables if t["table"]["name"] == "recipe_like")["event_triggers"] = [{
        "name": "recipe_like_notification",
        "definition": {"enable_manual": False, "insert": {"columns": "*"}},
        "retry_conf": {"num_retries": 3, "interval_sec": 10, "timeout_sec": 60},
        "webhook": "{{API_BASE_URL}}/v1/like-notification",
        "headers": headers,
    }]

    def action(name, path, args, output_type, roles):
        return {
            "name": name,
            "definition": {
                "kind": "synchronous", "type": "mutation",
                "handler": "{{API_BASE_URL}}/v1/" + path,
                "arguments": [{"name": key, "type": value} for key, value in args.items()],
                "output_type": output_type,
                "headers": headers,
                "forward_client_headers": False,
            },
            "permissions": [{"role": role} for role in roles],
        }

    actions = [
        action("login", "login", {"input": "LoginInput!"}, "LoginOutput!", ["anon", "user"]),
        action("signup", "signup", {"input": "SignupInput!"}, "SignupOutput!", ["anon", "user"]),
        action("Me", "me", {"input": "MeInput!"}, "MeOutput!", ["anon", "user"]),
        action("uploadRecipeImages", "upload-recipe-images", {"input": "UploadRecipeImagesInput!"}, "UploadOutput!", []),
        action("uploadUserImage", "upload-user-image", {"input": "UploadUserImageInput!"}, "UploadOutput!", []),
        action("initializePayment", "payment", {"recipe_id": "Int!", "amount": "numeric!"}, "InitializePaymentOutput!", []),
    ]
    return {
        "version": 3,
        "sources": [{
            "name": "default", "kind": "postgres", "tables": tables,
            "configuration": {"connection_info": {
                "database_url": {"from_env": "HASURA_GRAPHQL_DATABASE_URL"},
                "isolation_level": "read-committed", "use_prepared_statements": True,
                "pool_settings": {"max_connections": 10, "idle_timeout": 180, "retries": 1},
            }},
        }],
        "actions": actions,
        "custom_types": {
            "input_objects": [
                _input("LoginInput", email="String!", password="String!"),
                _input("SignupInput", email="String!", password="String!", first_name="String!", last_name="String!", username="String!"),
                _input("MeInput", token="String!"),
                _input("RecipeImageInput", image_name="String!", image_type="String!", image_base64str="String!", is_thumbnail="Boolean!"),
                _input("UploadRecipeImagesInput", recipe_id="Int!", recipe_images="[RecipeImageInput!]!"),
                _input("UserImageInput", image_name="String!", image_type="String!", image_base64str="String!"),
                _input("UploadUserImageInput", user_id="String", image="UserImageInput!"),
            ],
            "objects": [
                _input("AuthProfile", id="String!", email="String!", username="String!", FirstName="String!", LastName="String!", avatar_url="String!", bio="String!", created_at="String!", updated_at="String!"),
                _input("LoginOutput", user_id="String!", token="String!", loggedInUser="AuthProfile!"),
                _input("SignupOutput", user_id="String!", token="String!", createdUser="AuthProfile!"),
                _input("MeOutput", user_id="String!", token="String!", user="AuthProfile!"),
                _input("UploadOutput", success="Boolean!", image_url="String"),
                _input("InitializePaymentOutput", checkout_url="String!", tx_ref="String!", status="String!"),
            ],
            "scalars": [{"name": "numeric"}],
            "enums": [],
        },
    }


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", help="Write JSON to a file instead of stdout")
    args = parser.parse_args()
    result = json.dumps(build_metadata(), indent=2) + "\n"
    if args.output:
        with open(args.output, "w") as file:
            file.write(result)
    else:
        print(result, end="")
