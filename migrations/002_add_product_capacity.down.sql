ALTER TABLE order_items DROP CONSTRAINT IF EXISTS order_items_pkey;
ALTER TABLE order_items DROP COLUMN IF EXISTS capacity;
ALTER TABLE order_items ADD PRIMARY KEY (order_id, product_name);

ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_pkey;
ALTER TABLE cart_items DROP COLUMN IF EXISTS capacity;
ALTER TABLE cart_items ADD PRIMARY KEY (telegram_id, product_id);
