-- Enable pgcrypto extension for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Users table
CREATE TABLE IF NOT EXISTS "user" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    "FirstName" VARCHAR(255) NOT NULL DEFAULT '',
    "LastName" VARCHAR(255) NOT NULL DEFAULT '',
    avatar_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_email ON "user"(email);
CREATE INDEX IF NOT EXISTS idx_user_username ON "user"(username);

-- 2. Recipes table
CREATE TABLE IF NOT EXISTS recipe (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    ingredients TEXT NOT NULL DEFAULT '',
    instructions TEXT NOT NULL DEFAULT '',
    price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recipe_user_id ON recipe(user_id);

-- 3. Recipe images table
CREATE TABLE IF NOT EXISTS recipe_image (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id INT NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    img_url TEXT NOT NULL,
    is_thumbnail BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recipe_image_recipe_id ON recipe_image(recipe_id);

-- 4. Recipe likes table
CREATE TABLE IF NOT EXISTS recipe_like (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id INT NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_recipe_user_like UNIQUE (recipe_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_recipe_like_recipe_id ON recipe_like(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_like_user_id ON recipe_like(user_id);

-- 5. Transactions table (order initialization)
CREATE TABLE IF NOT EXISTS transaction (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trx_ref VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    recipe_id INT NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    amount DOUBLE PRECISION NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'ETB',
    checkout_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transaction_trx_ref ON transaction(trx_ref);
CREATE INDEX IF NOT EXISTS idx_transaction_user_id ON transaction(user_id);

-- 6. Payments table (completed / verified records)
CREATE TABLE IF NOT EXISTS payment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    recipe_id INT NOT NULL REFERENCES recipe(id) ON DELETE CASCADE,
    tx_ref VARCHAR(255) NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'ETB',
    chapa_checkout_url TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_tx_ref ON payment(tx_ref);
CREATE INDEX IF NOT EXISTS idx_payment_user_id ON payment(user_id);

-- 7. Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    initiator_id UUID NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_initiator_id ON notifications(initiator_id);
