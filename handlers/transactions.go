package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
)

const timezone = "Africa/Kampala"

type Transaction struct {
	ID        int32             `json:"id"`
	Products  []epharma.Product `json:"products"`
	CreatedAt time.Time         `json:"created_at"`
	UserID    int32             `json:"user_id"`
}

func Now() time.Time {
	t := time.Now()
	en, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}
	return t.In(en)
}

func groupTransactionsByDate(transactions []Transaction) map[string][]Transaction {
	transactionsByDate := make(map[string][]Transaction)
	for _, transaction := range transactions {
		date := transaction.CreatedAt.Format("2006-01-02")
		if _, ok := transactionsByDate[date]; !ok {
			transactionsByDate[date] = []Transaction{}
		}
		transactionsByDate[date] = append(transactionsByDate[date], transaction)

	}
	return transactionsByDate
}

func convertTransaction(transaction epharma.Transaction) Transaction {
	t := Transaction{
		ID:        transaction.ID,
		Products:  []epharma.Product{},
		CreatedAt: transaction.CreatedAt,
		UserID:    transaction.UserID,
	}

	err := json.Unmarshal(transaction.Items, &t.Products)
	if err != nil {
		log.Println(err)
	}
	return t
}

// ListTransactionsPaginated
func (h *Handlers) ListTransactionsPaginated(w http.ResponseWriter, r *http.Request) {
	page := egor.QueryInt(r, "page", 1)
	limit := egor.QueryInt(r, "limit", 10)

	offset := (page - 1) * limit
	transactions, err := h.Queries.ListTransactionsPaginated(r.Context(),
		epharma.ListTransactionsPaginatedParams{
			Offset: int32(offset),
			Limit:  int32(limit),
		})

	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// convert to a slice of transactions
	transactionsList := []Transaction{}
	for _, transaction := range transactions {
		transactionsList = append(transactionsList, convertTransaction(transaction))
	}

	egor.Render(w, r, "transactions/list", egor.Map{
		"transactions": groupTransactionsByDate(transactionsList),
	})
}

// CreateTransaction /POST
func (h *Handlers) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userId := egor.GetContextValue(r, "user").(epharma.User).ID

	type Payload struct {
		Products []epharma.Product `json:"products"`
	}

	var payload Payload
	err := egor.BodyParser(r, &payload)
	if err != nil {
		egor.SendJSONError(w, map[string]any{"error": err}, http.StatusBadRequest)
		return
	}

	// for each product, check if the quantity is available and update the stock
	for i := range payload.Products {
		product := payload.Products[i]
		stock, err := h.Queries.GetProduct(r.Context(), product.ID)
		if err != nil {
			egor.SendJSONError(w, map[string]any{"error": err}, http.StatusBadRequest)
			return
		}

		if stock.Quantity < product.Quantity {
			egor.SendJSONError(w, map[string]any{"error": "Insufficient stock"}, http.StatusUnprocessableEntity)
			return
		}

		product.ID = stock.ID
		product.Barcode = stock.Barcode
		product.GenericName = stock.GenericName
		product.BrandName = stock.BrandName
		product.CostPrice = stock.CostPrice
		product.SellingPrice = stock.SellingPrice
		product.CreatedAt = Now()
		payload.Products[i] = product
	}

	items, err := json.Marshal(payload.Products)
	if err != nil {
		egor.SendJSONError(w, map[string]any{"error": err}, http.StatusUnprocessableEntity)
		return
	}

	// Start transaction
	tx, err := h.Conn.Begin(r.Context())
	if err != nil {
		egor.SendJSONError(w, map[string]any{"error": "unable to init database transaction"})
		return
	}

	defer tx.Rollback(r.Context())
	qtx := h.Queries.WithTx(tx)

	// Reduce quantity for each product
	for _, product := range payload.Products {
		err = qtx.DecrementProduct(r.Context(), epharma.DecrementProductParams{
			ID:       product.ID,
			Quantity: product.Quantity,
		})
		if err != nil {
			egor.SendError(w, r, fmt.Errorf("decrement product quantity failed: %v", err), http.StatusNotFound)
			return
		}
	}

	transaction, err := qtx.CreateTransaction(r.Context(), epharma.CreateTransactionParams{
		UserID: userId,
		Items:  items,
	})

	if err != nil {
		egor.SendJSONError(w, map[string]any{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	// commit transaction
	tx.Commit(r.Context())

	// Return the JSON
	egor.SendJSON(w, convertTransaction(transaction))
}

// GetTransaction
func (h *Handlers) GetTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := egor.ParamInt(r, "id")
	transaction, err := h.Queries.GetTransaction(r.Context(), int32(transactionID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "transactions/detail", egor.Map{
		"transaction": convertTransaction(transaction),
	})
}

// DeleteTransaction
func (h *Handlers) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := egor.ParamInt(r, "id")
	tx, err := h.Conn.Begin(r.Context())
	if err != nil {
		egor.SendError(w, r, fmt.Errorf("unable to init database transaction: %v", err))
		return
	}
	defer tx.Rollback(r.Context())

	qtx := h.Queries.WithTx(tx)

	trans, err := qtx.GetTransaction(r.Context(), int32(transactionID))
	if err != nil {
		egor.SendError(w, r, fmt.Errorf("transaction not found: %v", err), http.StatusNotFound)
		return
	}

	// convert items from bytes (JSONB) to []epharma.Product
	transaction := convertTransaction(trans)

	// For each item, re-increment quantity
	for _, product := range transaction.Products {
		err = qtx.IncrementProduct(r.Context(), epharma.IncrementProductParams{
			ID:       product.ID,
			Quantity: product.Quantity,
		})
		if err != nil {
			egor.SendError(w, r, fmt.Errorf("re-increment product quantity failed: %v", err), http.StatusNotFound)
			return
		}
	}

	err = qtx.DeleteTransaction(r.Context(), int32(transactionID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// Commit the transaction
	tx.Commit(r.Context())
	egor.Redirect(w, r, "/transactions")
}
