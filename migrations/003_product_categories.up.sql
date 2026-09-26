DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'products'
          AND column_name = 'category'
    ) THEN
        RAISE EXCEPTION 'column products.category is missing; apply ALTER TABLE as database owner first';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM products WHERE category IN ('degreaser', 'solvent')
    ) THEN
        DELETE FROM cart_items;
        DELETE FROM products;
        INSERT INTO products (name, description, price, category) VALUES
            ('Strong Pro Обезжириватель универсальный', 'Быстро удаляет жир и масло с металла и пластика', 45000, 'degreaser'),
            ('Strong Pro Обезжириватель для металла', 'Для подготовки поверхности перед покраской', 52000, 'degreaser'),
            ('Strong Pro Обезжириватель промышленный', 'Концентрат для производственных линий', 68000, 'degreaser'),
            ('Strong Pro Растворитель 646', 'Универсальный растворитель для ЛКМ', 38000, 'solvent'),
            ('Strong Pro Уайт-спирит', 'Очищение инструмента и разбавление красок', 29000, 'solvent'),
            ('Strong Pro Растворитель быстросохнущий', 'Минимальный остаток после испарения', 41000, 'solvent');
    END IF;
END $$;
