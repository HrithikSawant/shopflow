-- Safe to run on a fresh database AND on one created from the old 001 schema.
-- (On a fresh database 001 already contains the CHECK constraints; this file
-- adds them when they are missing, and adds the outbox table.)

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_inventory_count_check') THEN
        ALTER TABLE products ADD CONSTRAINT products_inventory_count_check CHECK (inventory_count >= 0);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'order_items_quantity_check') THEN
        ALTER TABLE order_items ADD CONSTRAINT order_items_quantity_check CHECK (quantity > 0);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'order_items_price_check') THEN
        ALTER TABLE order_items ADD CONSTRAINT order_items_price_check CHECK (price >= 0);
    END IF;
END $$;

ALTER TABLE order_items ALTER COLUMN order_id SET NOT NULL;
ALTER TABLE order_items ALTER COLUMN product_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS order_items_order_id_idx ON order_items (order_id);

-- Transactional outbox: written in the same statement as the order, published
-- to RabbitMQ by the API's relay.
CREATE TABLE IF NOT EXISTS outbox_events (
    id           BIGSERIAL PRIMARY KEY,
    order_id     BIGINT NOT NULL REFERENCES orders(id),
    payload      JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS outbox_events_unpublished_idx
    ON outbox_events (id) WHERE published_at IS NULL;
