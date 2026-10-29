#!/usr/bin/env python3
"""Add optional, repeatable demo recipes through Hasura's authenticated SQL API.

Requires HASURA_URL and HASURA_GRAPHQL_ADMIN_SECRET, like bootstrap.py.
Refreshes only recognized original sample dishes and placeholder photographs.
Does not change application metadata, unrelated recipes, or existing passwords.
"""
import sys

from bootstrap import DeploymentError, Hasura, ROOT


def seed():
    client = Hasura()
    client.sql((ROOT / "deploy/database/seed.sql").read_text())
    client.sql((ROOT / "deploy/database/refresh_seed_photography.sql").read_text())
    result = client.sql("SELECT count(*)::text FROM public._food_recipe_seed", True)
    count = int(result["result"][1][0])
    print(f"Sample seed complete: {count} recipes. Existing data preserved.")
    return count


if __name__ == "__main__":
    try:
        seed()
    except DeploymentError as error:
        print("Seeding blocked: " + str(error), file=sys.stderr)
        raise SystemExit(1)
