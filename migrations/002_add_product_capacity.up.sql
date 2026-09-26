ALTER TABLE cart_items ADD COLUMN IF NOT EXISTS capacity TEXT NOT NULL DEFAULT 'не указана';
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_pkey;
ALTER TABLE cart_items ADD PRIMARY KEY (telegram_id, product_id, capacity);

ALTER TABLE order_items ADD COLUMN IF NOT EXISTS capacity TEXT NOT NULL DEFAULT 'не указана';
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_pkey;
ALTER TABLE order_items ADD PRIMARY KEY (order_id, product_name, capacity);
