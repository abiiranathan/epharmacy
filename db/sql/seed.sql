INSERT INTO
    products (generic_name, brand_name, quantity, cost_price, selling_price, expiry_date, barcode)
SELECT
    md5(random()::text),
    md5(random()::text),
    (random() * 100)::int,
    (random() * 100)::numeric(10, 2),
    (random() * 100)::numeric(10, 2),
    CURRENT_DATE + (random() * 365)::int,
    md5(random()::text)
FROM
    generate_series(1, 100) s;