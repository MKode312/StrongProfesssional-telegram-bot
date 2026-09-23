CREATE TABLE users (
    telegram_id BIGINT PRIMARY KEY,
    email TEXT NOT NULL,
    phone TEXT NOT NULL,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price BIGINT NOT NULL CHECK (price >= 0),
    category TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE cart_items (
    telegram_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (telegram_id, product_id)
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL REFERENCES users(telegram_id),
    total BIGINT NOT NULL CHECK (total >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id BIGINT,
    product_name TEXT NOT NULL,
    unit_price BIGINT NOT NULL CHECK (unit_price >= 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, product_name)
);

INSERT INTO products (name, description, price, category) VALUES
    ('Strong Pro Обезжириватель универсальный', 'Быстро удаляет жир и масло с металла и пластика', 45000, 'degreaser'),
    ('Strong Pro Обезжириватель для металла', 'Для подготовки поверхности перед покраской', 52000, 'degreaser'),
    ('Strong Pro Обезжириватель промышленный', 'Концентрат для производственных линий', 68000, 'degreaser'),
    ('Strong Pro Растворитель 646', 'Универсальный растворитель для ЛКМ', 38000, 'solvent'),
    ('Strong Pro Уайт-спирит', 'Очищение инструмента и разбавление красок', 29000, 'solvent'),
    ('Strong Pro Растворитель быстросохнущий', 'Минимальный остаток после испарения', 41000, 'solvent');
