// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package epharma

import (
	"time"

	"github.com/abiiranathan/dbtypes"
)

type Invoice struct {
	ID            int32        `json:"id"`
	InvoiceNumber string       `json:"invoice_number"`
	PurchaseDate  dbtypes.Date `json:"purchase_date"`
	InvoiceTotal  float64      `json:"invoice_total"`
	AmountPaid    float64      `json:"amount_paid"`
	Balance       float64      `json:"balance"`
	Supplier      string       `json:"supplier"`
	UserID        int32        `json:"user_id"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Product struct {
	ID           int32          `json:"id"`
	GenericName  string         `json:"generic_name"`
	BrandName    string         `json:"brand_name"`
	Quantity     int32          `json:"quantity"`
	CostPrice    float64        `json:"cost_price"`
	SellingPrice float64        `json:"selling_price"`
	ExpiryDates  []dbtypes.Date `json:"expiry_dates"`
	Barcode      string         `json:"barcode"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type StockIn struct {
	ID         int32        `json:"id"`
	ProductID  int32        `json:"product_id"`
	InvoiceID  int32        `json:"invoice_id"`
	Quantity   int32        `json:"quantity"`
	CostPrice  float64      `json:"cost_price"`
	ExpiryDate dbtypes.Date `json:"expiry_date"`
	Comment    string       `json:"comment"`
	CreatedAt  time.Time    `json:"created_at"`
}

type Transaction struct {
	ID        int32     `json:"id"`
	Items     []byte    `json:"items"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int32     `json:"user_id"`
}

type User struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}