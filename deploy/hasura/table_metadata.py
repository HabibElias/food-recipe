"""Hasura v2 table metadata for the deployed Nuxt/Go application.

Permissions use database column names, including columns with GraphQL aliases.
Admin-only adapters retain the original singular SQL and GraphQL root names.
"""

SESSION_USER = "X-Hasura-User-Id"
OWN_ROW = {"user_id": {"_eq": SESSION_USER}}
OWN_RECIPE = {"recipe": OWN_ROW}


def _table(name):
    return {"schema": "public", "name": name}


def _object(name, column):
    return {"name": name, "using": {"foreign_key_constraint_on": column}}


def _array(name, table, column):
    return {
        "name": name,
        "using": {"foreign_key_constraint_on": {"table": _table(table), "column": column}},
    }


def _select(role, columns, row_filter=None, aggregate=True):
    return {
        "role": role,
        "permission": {
            "columns": columns,
            "filter": row_filter or {},
            "allow_aggregations": aggregate,
        },
    }


def _insert(columns, check):
    return {"role": "user", "permission": {"columns": columns, "check": check}}


def _update(columns, row_filter):
    return {
        "role": "user",
        "permission": {"columns": columns, "filter": row_filter, "check": row_filter},
    }


def _delete(row_filter):
    return {"role": "user", "permission": {"filter": row_filter}}


def build_tables():
    """Return metadata entries for every application table, without secret values."""
    tables = {}

    def add(name, columns, *, objects=(), arrays=(), public=True, user_filter=None):
        entry = {"table": _table(name)}
        if objects:
            entry["object_relationships"] = [_object(*obj) for obj in objects]
        if arrays:
            entry["array_relationships"] = [_array(*arr) for arr in arrays]
        selects = []
        if public:
            selects.append(_select("anon", columns))
        selects.append(_select("user", columns, user_filter))
        entry["select_permissions"] = selects
        tables[name] = entry
        return entry

    # Login/signup/me return private account data from custom action output types.
    # Never make password hashes or email addresses selectable through a profile.
    add("user", ["id", "username", "FirstName", "LastName", "avatar_url", "bio", "created_at", "updated_at"],
        arrays=[("recipes", "recipe", "user_id")])
    add("category", ["id", "category_name", "category_description"])

    recipe = add("recipe", ["id", "user_id", "title", "description", "price", "created_at", "updated_at", "category_id", "is_paid", "prep_time"],
                 objects=[("user", "user_id"), ("category", "category_id")],
                 arrays=[("recipe_images", "recipe_image", "recipe_id"),
                         ("recipe_likes", "recipe_like", "recipe_id"),
                         ("recipe_steps", "recipe_step", "recipe_id"),
                         ("recipe_ingredients", "recipe_ingredient", "recipe_id"),
                         ("recipe_ratings", "recipe_rating", "recipe_id"),
                         ("recipe_comments", "recipe_comment", "recipe_id"),
                         ("bookmarks", "bookmark", "recipe_id"),
                         ("payments", "payment", "recipe_id")])
    recipe["insert_permissions"] = [_insert(
        ["user_id", "title", "description", "price", "category_id", "is_paid", "prep_time"], OWN_ROW)]
    recipe["update_permissions"] = [_update(
        ["title", "description", "price", "category_id", "is_paid", "prep_time"], OWN_ROW)]
    recipe["delete_permissions"] = [_delete(OWN_ROW)]

    image = add("recipe_image", ["id", "recipe_id", "img_url", "is_thumbnail", "created_at"],
                objects=[("recipe", "recipe_id")])
    image["insert_permissions"] = [_insert(["recipe_id", "img_url", "is_thumbnail"], OWN_RECIPE)]
    image["update_permissions"] = [_update(["is_thumbnail"], OWN_RECIPE)]
    image["delete_permissions"] = [_delete(OWN_RECIPE)]

    like = add("recipe_like", ["id", "recipe_id", "user_id", "created_at"],
               objects=[("recipe", "recipe_id"), ("user", "user_id")])
    like["insert_permissions"] = [_insert(["recipe_id", "user_id"], OWN_ROW)]
    like["delete_permissions"] = [_delete(OWN_ROW)]

    # The old free-text ingredients/instructions columns are deliberately excluded
    # from recipe permissions so paid content cannot bypass these row filters.
    paid_access = {"_or": [
        {"is_paid": {"_eq": False}}, OWN_ROW,
        {"payments": {"_and": [OWN_ROW, {"status": {"_eq": "success"}}]}},
    ]}
    for name, content_columns in [
        ("recipe_step", ["step_no", "step_description"]),
        ("recipe_ingredient", ["ingredient"]),
    ]:
        cols = ["id", "recipe_id"] + content_columns
        child = add(name, cols, objects=[("recipe", "recipe_id")], public=False,
                    user_filter={"recipe": paid_access})
        child["select_permissions"].insert(0, _select("anon", cols, {"recipe": {"is_paid": {"_eq": False}}}))
        child["insert_permissions"] = [_insert(["recipe_id"] + content_columns, OWN_RECIPE)]
        child["update_permissions"] = [_update(content_columns, OWN_RECIPE)]
        child["delete_permissions"] = [_delete(OWN_RECIPE)]

    rating = add("recipe_rating", ["id", "recipe_id", "user_id", "rating", "created_at"],
                 objects=[("recipe", "recipe_id"), ("user", "user_id")])
    rating["insert_permissions"] = [_insert(["recipe_id", "user_id", "rating"], OWN_ROW)]
    rating["update_permissions"] = [_update(["rating"], OWN_ROW)]
    rating["delete_permissions"] = [_delete(OWN_ROW)]

    comment = add("recipe_comment", ["id", "recipe_id", "user_id", "comment_body", "created_at"],
                  objects=[("recipe", "recipe_id"), ("user", "user_id")])
    comment["insert_permissions"] = [_insert(["recipe_id", "user_id", "comment_body"], OWN_ROW)]
    comment["update_permissions"] = [_update(["comment_body"], OWN_ROW)]
    comment["delete_permissions"] = [_delete(OWN_ROW)]

    bookmark = add("bookmark", ["id", "recipe_id", "user_id", "created_at"],
                   objects=[("recipe", "recipe_id"), ("user", "user_id")], public=False,
                   user_filter=OWN_ROW)
    bookmark["insert_permissions"] = [_insert(["recipe_id", "user_id"], OWN_ROW)]
    bookmark["delete_permissions"] = [_delete(OWN_ROW)]

    add("transaction", ["id", "trx_ref", "user_id", "recipe_id", "amount", "currency", "checkout_url", "created_at"],
        objects=[("recipe", "recipe_id"), ("user", "user_id")], public=False, user_filter=OWN_ROW)
    add("payment", ["id", "user_id", "recipe_id", "status", "created_at"],
        objects=[("recipe", "recipe_id"), ("user", "user_id")], public=False,
        user_filter={"_and": [{"status": {"_eq": "success"}}, {"_or": [OWN_ROW, OWN_RECIPE]}]})

    notification = add("notifications", ["id", "user_id", "initiator_id", "message", "is_read", "created_at"],
                       objects=[("user", "user_id"), ("initiator", "initiator_id")],
                       public=False, user_filter=OWN_ROW)
    notification["configuration"] = {"column_config": {"is_read": {"custom_name": "read"}}}
    notification["update_permissions"] = [_update(["is_read"], OWN_ROW)]

    return list(tables.values())
