-- name: ListUsers :many
SELECT * FROM users;

-- name: CreateUser :one
INSERT INTO
    users (username, password)
VALUES
    ($1, $2) RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: UpdateUser :exec
-- Only update password if it's not empty
UPDATE users SET username = sqlc.arg('username'), 
password= CASE WHEN  @update_password::bool
THEN @password
    ELSE password
    END
WHERE id = sqlc.arg('id');

-- name: ActivateUser :exec
UPDATE users SET is_active = TRUE WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users SET is_active = FALSE WHERE id = $1;

-- name: PromoteUser :exec
UPDATE users SET is_admin = TRUE WHERE id = $1;

-- name: DemoteUser :exec
UPDATE users SET is_admin = FALSE WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;


-- -- Product queries ----------------

-- name: ListProductsPaginated :many
-- Filter by name if provided.
SELECT * FROM products WHERE 
CASE WHEN @name::text != ''
    THEN generic_name ILIKE '%' || @name::text || '%' OR brand_name ILIKE '%' || @name::text || '%'
    ELSE TRUE
END
ORDER BY id LIMIT @lim OFFSET @off;

-- name: CreateProduct :one
INSERT INTO
    products (generic_name, brand_name, quantity, 
    cost_price, selling_price, barcode, expiry_dates)
VALUES
    ($1, $2, $3, $4, $5, $6, $7) RETURNING *;


-- name: CreateProducts :copyfrom
INSERT INTO
    products (generic_name, brand_name, quantity, 
    cost_price, selling_price, barcode, expiry_dates)
VALUES
    ($1, $2, $3, $4, $5, $6, $7);


-- name: GetProduct :one
SELECT * FROM products WHERE id = $1;

-- name: UpdateProduct :exec
UPDATE products SET generic_name = $1, brand_name = $2, 
    quantity = $3, cost_price = $4, selling_price = $5, 
    barcode = $6, expiry_dates=$7 WHERE id = $8;

-- name: IncrementProduct :exec
UPDATE products SET quantity=quantity + $2 WHERE id = $1;

-- name: DecrementProduct :exec
UPDATE products SET quantity=quantity - $2 WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM products WHERE id = $1;

-- name: GetProductByBarcode :one
SELECT * FROM products WHERE barcode = $1;

-- Full text search for the products table
-- name: SearchProducts :many
SELECT * FROM products
WHERE
    generic_name ILIKE '%' || @name::text || '%'
    OR brand_name ILIKE '%' || @name::text || '%'
ORDER BY id LIMIT 50;

-- name: CountProducts :one
SELECT COUNT(*) AS count FROM products;

-- -- Transactions queries ----------------
-- name: ListTransactionsPaginated :many
SELECT * FROM transactions ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateTransaction :one
INSERT INTO
    transactions (items, user_id)
VALUES
    ($1, $2) RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions WHERE id = $1;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1;

-- Ruturn 10 most common products in transactions
-- order by count
-- name: MostCommonProducts :many
SELECT DISTINCT * FROM products p
JOIN (
    SELECT DISTINCT (item->>'id')::int AS product_id,
           COUNT(*) AS count
    FROM transactions
    CROSS JOIN LATERAL jsonb_array_elements(items) item
    GROUP BY product_id
) t ON p.id = t.product_id
ORDER BY t.count DESC
LIMIT $1;

-- -- Invoices queries ----------------

-- name: ListInvoicesPaginated :many
SELECT * FROM invoices ORDER BY id LIMIT $1 OFFSET $2;

-- name: CreateInvoice :one
INSERT INTO
    invoices (invoice_number, purchase_date, invoice_total,
     amount_paid, supplier, user_id)
VALUES
    ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetInvoice :one
SELECT * FROM invoices WHERE id = $1;

-- name: UpdateInvoice :exec
UPDATE invoices SET 
        invoice_number = $1, 
        purchase_date = $2, 
        invoice_total = $3, 
        amount_paid = $4, 
        supplier = $5, 
        user_id = $6 
        WHERE id = $7;

-- name: DeleteInvoice :exec
DELETE FROM invoices WHERE id = $1;

-- name: SearchInvoices :many
SELECT * FROM invoices WHERE invoice_number = $1;

-- name: GetInvoiceByNumber :one
SELECT * FROM invoices WHERE invoice_number = $1;

-- ================== StockIN Queries =========================
-- name: InvoiceItems :many
SELECT stock_in.*, 
    products.generic_name, products.brand_name
FROM stock_in
JOIN products ON stock_in.product_id = products.id
WHERE stock_in.invoice_id = (SELECT id FROM invoices WHERE invoice_number=$1 LIMIT 1)
ORDER BY stock_in.id;

-- name: AddProductToInvoice :exec
INSERT INTO stock_in (product_id, invoice_id, quantity, cost_price, expiry_date, comment)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DeleteStockIn :exec
DELETE FROM stock_in WHERE id = $1;

-- name: IncrementProductQuantity :exec
UPDATE products SET quantity = quantity + $1 WHERE id = $2;

-- name: DecrementProductQuantity :exec
UPDATE products SET quantity = quantity - $1 WHERE id = $2 AND quantity >= $1;

-- name: ReplaceProductExpiry :exec
UPDATE products SET expiry_dates = $1 WHERE id = $2;

-- Add a new expiry date to the product expiry dates.
-- Do this only if this expiry date is not already in the list.
-- name: AddProductExpiry :exec
UPDATE products
SET expiry_dates = COALESCE(expiry_dates, ARRAY[]::date[]) || @expiry_date::date
WHERE id = @id::int
  AND NOT (expiry_dates @> ARRAY[@expiry_date::date])
  AND @expiry_date::date IS NOT NULL;

-- name: RemoveProductExpiry :exec
UPDATE products
SET expiry_dates = array_remove(expiry_dates, @expiry_date::date)
WHERE id = @id::int;


-- name: GetStockIn :one
SELECT * FROM stock_in WHERE id = $1;


