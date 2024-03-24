-- This table will store the stock balances for each product per day.
-- It is updated the once a day, before the first transaction of the day.
-- It will also be updated after each stock in transaction.
CREATE TABLE IF NOT EXISTS stock_balances (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL,
    opening_quantity INTEGER NOT NULL DEFAULT 0,
    quantity_in INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- FOREIGN KEYS
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);


CREATE OR REPLACE FUNCTION update_stock_balance()
RETURNS TRIGGER AS $$
DECLARE
    item_id int; 
    prod_quantity int;
BEGIN
    -- Iterate over each item in the transactions.items array
    FOR item_id IN SELECT (jsonb_array_elements(NEW.items)->>'id')::INTEGER
    LOOP
        -- Check if the item's product_id already has a stock balance record for the current date
        IF NOT EXISTS (
            SELECT 1 FROM stock_balances
            WHERE product_id = item_id
            AND DATE_TRUNC('day', created_at) = DATE_TRUNC('day', CURRENT_TIMESTAMP)
        ) THEN
            -- Select the current stock balance for the product
            SELECT products.quantity INTO prod_quantity FROM products WHERE id = item_id;
            
            -- Create a new stock balance record for the product for the current date
            INSERT INTO stock_balances (product_id, opening_quantity, quantity_in)
            VALUES (item_id, prod_quantity, 0);
        END IF;
    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a trigger to update the stock balance before each transaction
CREATE TRIGGER update_stock_balance_trigger
BEFORE INSERT ON transactions
FOR EACH ROW
EXECUTE FUNCTION update_stock_balance();


-- Create a trigger to update the stock balance after each insert into stock_in
CREATE OR REPLACE FUNCTION update_stock_balance_after_stock_in()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the product already has a stock balance record for the current date
    IF NOT EXISTS (
        SELECT 1 FROM stock_balances
        WHERE product_id = NEW.product_id
        AND DATE_TRUNC('day', created_at) = DATE_TRUNC('day', CURRENT_TIMESTAMP)
    ) THEN
        -- Create a new stock balance record for the product for the current date
        INSERT INTO stock_balances (product_id, opening_quantity, quantity_in)
        VALUES (NEW.product_id, 0, NEW.quantity);
    ELSE
        -- Update the quantity_in column of the stock balance record for the product for the current date
        UPDATE stock_balances
        SET quantity_in = quantity_in + NEW.quantity
        WHERE product_id = NEW.product_id
        AND DATE_TRUNC('day', created_at) = DATE_TRUNC('day', CURRENT_TIMESTAMP);
    END IF;

    RETURN NEW;
END;

$$ LANGUAGE plpgsql;

CREATE TRIGGER update_stock_balance_after_stock_in_trigger
BEFORE INSERT ON stock_in
FOR EACH ROW
EXECUTE FUNCTION update_stock_balance_after_stock_in();


-- Create a trigger that initializes stock_balances when a product is inserted
CREATE OR REPLACE FUNCTION initialize_stock_balance()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO stock_balances (product_id, opening_quantity, quantity_in)
    VALUES (NEW.id, 0, NEW.quantity);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER initialize_stock_balance_trigger
AFTER INSERT ON products
FOR EACH ROW
EXECUTE FUNCTION initialize_stock_balance();


-- Now create a stock_card view that will show the stock balances for each product per day
CREATE OR REPLACE VIEW stock_card AS
WITH QuantityOutCTE AS (
    SELECT 
        created_at::date AS date,
        (item->>'id')::INTEGER AS product_id,
        SUM((item->>'quantity')::INTEGER) AS total_quantity_out
    FROM transactions,
        LATERAL jsonb_array_elements(items) AS items(item)
    WHERE (item->>'id')::INTEGER IN (SELECT id FROM products)
    GROUP BY date, product_id
)

SELECT 
    stock_balances.created_at::date,
    products.id AS product_id,
    products.generic_name,
    products.brand_name,
    stock_balances.opening_quantity,
    stock_balances.quantity_in,
    COALESCE(total_quantity_out, 0) AS quantity_out,
    stock_balances.opening_quantity + stock_balances.quantity_in - COALESCE(total_quantity_out, 0) AS closing_quantity
FROM stock_balances
JOIN products ON stock_balances.product_id = products.id
LEFT JOIN QuantityOutCTE ON stock_balances.product_id = QuantityOutCTE.product_id 
AND stock_balances.created_at::date = QuantityOutCTE.date
ORDER BY stock_balances.created_at DESC, products.id;

