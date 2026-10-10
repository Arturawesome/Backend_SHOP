CREATE TABLE IF NOT EXISTS product (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    price BIGINT NOT NULL CHECK (price >= 0),
    available_stock INTEGER NOT NULL CHECK (available_stock >= 0),
    last_update_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    supplier_id UUID NOT NULL REFERENCES supplier(id),
    image_id UUID REFERENCES image(id) ON DELETE SET NULL
);
