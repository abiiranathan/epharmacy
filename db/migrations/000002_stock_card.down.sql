DROP TRIGGER update_stock_balance_after_stock_in_trigger ON stock_in;
DROP FUNCTION update_stock_balance_after_stock_in;

DROP TRIGGER update_stock_balance_trigger ON transactions;
DROP FUNCTION update_stock_balance;

DROP TRIGGER initialize_stock_balance_trigger ON products;
DROP FUNCTION initialize_stock_balance;

DROP TABLE IF EXISTS stock_balances;