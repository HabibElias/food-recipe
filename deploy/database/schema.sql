-- Apply after ../../init.sql. These additions match the checked-in Nuxt queries.
-- The deployment initializer only applies this to an explicitly new database.
-- Reapplying it preserves existing rows.
BEGIN;

ALTER TABLE "user" ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS category (
    id SERIAL PRIMARY KEY,
    category_name TEXT NOT NULL UNIQUE,
    category_description TEXT NOT NULL DEFAULT ''
);
INSERT INTO category (category_name, category_description)
VALUES ('General', 'Recipes from every kitchen')
ON CONFLICT (category_name) DO NOTHING;

ALTER TABLE recipe ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES category(id);
ALTER TABLE recipe ADD COLUMN IF NOT EXISTS is_paid BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE recipe ADD COLUMN IF NOT EXISTS prep_time INTEGER NOT NULL DEFAULT 0;
UPDATE recipe SET category_id = (SELECT id FROM category WHERE category_name = 'General')
WHERE category_id IS NULL;
ALTER TABLE recipe ALTER COLUMN category_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_recipe_category_id ON recipe(category_id);

CREATE TABLE IF NOT EXISTS recipe_step (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    step_no INTEGER NOT NULL CHECK (step_no > 0),
    step_description TEXT NOT NULL,
    UNIQUE (recipe_id, step_no)
);
CREATE INDEX IF NOT EXISTS idx_recipe_step_recipe_id ON recipe_step(recipe_id);

CREATE TABLE IF NOT EXISTS recipe_ingredient (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    ingredient TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredient_recipe_id ON recipe_ingredient(recipe_id);

CREATE TABLE IF NOT EXISTS recipe_rating (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (recipe_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_recipe_rating_recipe_id ON recipe_rating(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_rating_user_id ON recipe_rating(user_id);

CREATE TABLE IF NOT EXISTS recipe_comment (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    comment_body TEXT NOT NULL CHECK (length(btrim(comment_body)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_recipe_comment_recipe_id ON recipe_comment(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_comment_user_id ON recipe_comment(user_id);

CREATE TABLE IF NOT EXISTS bookmark (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (recipe_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_bookmark_recipe_id ON bookmark(recipe_id);
CREATE INDEX IF NOT EXISTS idx_bookmark_user_id ON bookmark(user_id);

COMMIT;
