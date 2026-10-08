CREATE TABLE IF NOT EXISTS users (
 id BIGSERIAL PRIMARY KEY, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL,
 role TEXT NOT NULL DEFAULT 'customer' CHECK (role IN ('customer','admin')), created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS products (
 id BIGSERIAL PRIMARY KEY, title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
 price NUMERIC(12,2) NOT NULL CHECK (price >= 0), rating DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (rating BETWEEN 0 AND 5),
 images TEXT[] NOT NULL CHECK (cardinality(images) >= 3), stock_quantity INTEGER NOT NULL CHECK (stock_quantity >= 0),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS promo_codes (
 id BIGSERIAL PRIMARY KEY, code TEXT NOT NULL UNIQUE, discount_percentage INTEGER NOT NULL CHECK (discount_percentage BETWEEN 0 AND 100), is_active BOOLEAN NOT NULL DEFAULT true
);
CREATE TABLE IF NOT EXISTS carts (
 user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, promo_id BIGINT REFERENCES promo_codes(id) ON DELETE SET NULL
);
CREATE TABLE IF NOT EXISTS cart_items (
 id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES carts(user_id) ON DELETE CASCADE,
 product_id BIGINT NOT NULL REFERENCES products(id), quantity INTEGER NOT NULL CHECK (quantity > 0), UNIQUE(user_id,product_id)
);
INSERT INTO promo_codes (code, discount_percentage) VALUES ('SAVE10',10) ON CONFLICT (code) DO NOTHING;
