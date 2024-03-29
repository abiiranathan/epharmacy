CREATE VIEW product_aggregates AS
SELECT
    t.created_at::date AS transaction_date,
    (item.product->>'id')::int AS product_id,
    SUM((item.product->>'quantity')::int) AS quantity_sold,
    SUM((item.product->>'quantity')::int * (item.product->>'selling_price')::numeric) AS income
FROM
    transactions t
CROSS JOIN LATERAL jsonb_array_elements(t.items) AS item(product)
GROUP BY
    t.created_at::date,
    (item.product->>'id')::int;

CREATE VIEW product_sales AS
WITH product_info AS (
    SELECT
        id,
        generic_name AS product_name,
        cost_price,
        selling_price
    FROM
        products
)
SELECT
    pa.transaction_date,
    pa.product_id,
    pi.product_name,
    pi.cost_price,
    pi.selling_price,
    pa.quantity_sold,
    pa.income::double precision AS income,
    (pi.selling_price * pa.quantity_sold - pi.cost_price * pa.quantity_sold)::double precision AS profit
FROM
    product_aggregates pa
JOIN
    product_info pi ON pa.product_id = pi.id
ORDER BY
    pa.transaction_date DESC,
    pa.product_id;



-- Aggregate sales report. This is a daily summary of the sales.
CREATE VIEW sales_reports AS
SELECT
    transaction_date,
    SUM(income)::double precision AS total_income
FROM
    product_sales
GROUP BY
    transaction_date
ORDER BY
    transaction_date DESC;

