
-- Track daily opening and closing balances
CREATE VIEW stock_cards AS
WITH DispenseLogsAggregated AS (
    SELECT
        billables.item_id,
        DATE(dispense_logs.dispensed_at) AS dispense_date,
        SUM(dispense_logs.quantity_issued) AS total_quantity_issued
    FROM dispense_logs
    INNER JOIN billables ON billables.id = dispense_logs.billable_id
    GROUP BY billables.item_id, DATE(dispense_logs.dispensed_at)
    -- Add union to include internal_issued_items
    UNION ALL
    SELECT
        internal_issued_items.item_id,
        DATE(internal_issued_items.created_at) AS dispense_date,
        SUM(internal_issued_items.quantity_taken) AS total_quantity_issued
    FROM internal_issued_items
    GROUP BY internal_issued_items.item_id, DATE(internal_issued_items.created_at)
),
StockCardCTE AS (
    SELECT
        DATE(stock_balances.created_at) AS date,
        stock_balances.item_id,
        inventory_items.name AS item_name,
        stock_balances.opening_quantity,
        stock_balances.quantity_in,
        COALESCE(dla.total_quantity_issued, 0) AS quantity_out,
        (stock_balances.opening_quantity + stock_balances.quantity_in - COALESCE(dla.total_quantity_issued, 0)) AS closing_balance
    FROM inventory_items
    -- Use LEFT JOIN to include items with no stock balances
    JOIN stock_balances ON inventory_items.id = stock_balances.item_id
    LEFT JOIN DispenseLogsAggregated dla ON inventory_items.id = dla.item_id
        AND DATE(stock_balances.created_at) = dla.dispense_date
)
SELECT *
FROM StockCardCTE
ORDER BY date DESC, item_name ASC;
