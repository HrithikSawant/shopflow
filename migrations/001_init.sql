CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL,
    inventory_count INTEGER NOT NULL CHECK (inventory_count >= 0),
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'PENDING',
    total_amount NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id),
    product_id BIGINT REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price NUMERIC(10,2) NOT NULL CHECK (price >= 0)
);

INSERT INTO products (name, description, price, inventory_count)
SELECT v.name, v.description, v.price, v.inventory_count
FROM (VALUES
    ('Mechanical Keyboard', 'Hot swappable keyboard', 4999, 25),
    ('Gaming Mouse', 'Wireless mouse', 2999, 50)
) AS v(name, description, price, inventory_count)
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.name = v.name);
