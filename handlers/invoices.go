package handlers

import (
	"fmt"
	"net/http"

	"github.com/abiiranathan/dbtypes"
	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
)

// ListInvoicesPaginated
func (h *Handlers) ListInvoicesPaginated(w http.ResponseWriter, r *http.Request) {
	page := min(egor.QueryInt(r, "page", 1), 1)
	limit := max(egor.QueryInt(r, "limit", 10), 10)

	offset := (page - 1) * limit

	invoices, err := h.Queries.ListInvoicesPaginated(r.Context(), epharma.ListInvoicesPaginatedParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "invoices/list", egor.Map{
		"invoices": invoices,
	})
}

// RenderInvoiceCreatePage
func (h *Handlers) RenderInvoiceCreatePage(w http.ResponseWriter, r *http.Request) {
	egor.Render(w, r, "invoices/create", egor.Map{})
}

// CreateInvoice
func (h *Handlers) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var params epharma.CreateInvoiceParams
	err := egor.BodyParser(r, &params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// validate invoice
	if params.InvoiceNumber == "" || params.InvoiceTotal == 0 || params.AmountPaid == 0 {
		egor.SendError(w, r, fmt.Errorf("invalid payload"), http.StatusBadRequest)
		return
	}

	// Add user id
	user := egor.GetContextValue(r, "user").(epharma.User)
	params.UserID = user.ID

	invoice, err := h.Queries.CreateInvoice(r.Context(), params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// Allow user add items to invoice.
	egor.Redirect(w, r, fmt.Sprintf("/invoices/view/%d", invoice.ID))
}

// GetInvoice
func (h *Handlers) GetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := egor.ParamInt(r, "id")
	invoice, err := h.Queries.GetInvoice(r.Context(), int32(invoiceID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	invoiceItems, err := h.Queries.InvoiceItems(r.Context(), invoice.InvoiceNumber)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "invoices/view", egor.Map{
		"invoice":      invoice,
		"invoiceItems": invoiceItems,
	})
}

// RenderInvoiceUpdatePage
func (h *Handlers) RenderInvoiceUpdatePage(w http.ResponseWriter, r *http.Request) {
	invoiceID := egor.ParamInt(r, "id")
	invoice, err := h.Queries.GetInvoice(r.Context(), int32(invoiceID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "invoices/update", egor.Map{
		"invoice": invoice,
	})
}

// UpdateInvoice
func (h *Handlers) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := egor.ParamInt(r, "id")
	var params epharma.UpdateInvoiceParams
	err := egor.BodyParser(r, &params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	params.ID = int32(invoiceID)

	err = h.Queries.UpdateInvoice(r.Context(), params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// redirect to invoice details page.
	egor.Redirect(w, r, fmt.Sprintf("/invoices/view/%d", invoiceID))
}

// DeleteInvoice
func (h *Handlers) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := egor.ParamInt(r, "id")
	err := h.Queries.DeleteInvoice(r.Context(), int32(invoiceID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Redirect(w, r, "/invoices")
}

// GetInvoiceByNumber
func (h *Handlers) GetInvoiceByNumber(w http.ResponseWriter, r *http.Request) {
	invoiceNumber := egor.Query(r, "invoice_number")
	invoice, err := h.Queries.GetInvoiceByNumber(r.Context(), invoiceNumber)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.SendJSON(w, invoice)
}

// ListInvoiceProducts
func (h *Handlers) ListInvoiceProducts(w http.ResponseWriter, r *http.Request) {
	invoiceNumber := egor.Query(r, "invoice_number")
	if invoiceNumber == "" {
		egor.SendError(w, r, fmt.Errorf("no provided invoice number"), http.StatusBadRequest)
		return
	}

	products, err := h.Queries.InvoiceItems(r.Context(), invoiceNumber)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.SendJSON(w, products)
}

func (h *Handlers) NewStockIn(w http.ResponseWriter, r *http.Request) {
	var stockin epharma.AddProductToInvoiceParams
	err := egor.BodyParser(r, &stockin)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// validate it
	if stockin.Quantity <= 0 || stockin.InvoiceID <= 0 && stockin.ProductID <= 0 || stockin.CostPrice <= 0 {
		egor.SendError(w, r, fmt.Errorf("all fields are required"), http.StatusBadRequest)
		return
	}

	tx, err := h.Conn.Begin(r.Context())
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := h.Queries.WithTx(tx)

	// New stock in
	err = qtx.AddProductToInvoice(r.Context(), stockin)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// If out of stock, replace expiry date
	product, err := h.Queries.GetProduct(r.Context(), stockin.ProductID)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// If product is out of stock, replace expiry date with new one
	// if expiry date for the new stock in is set.
	if product.Quantity <= 0 && !stockin.ExpiryDate.IsZero() {
		err = qtx.ReplaceProductExpiry(r.Context(), epharma.ReplaceProductExpiryParams{
			ID:          stockin.ProductID,
			ExpiryDates: []dbtypes.Date{stockin.ExpiryDate},
		})
	} else if product.Quantity > 0 && !stockin.ExpiryDate.IsZero() {
		// Add a new expiry date to the product expiry dates.
		// Do this only if this expiry date is not already in the list.
		err = qtx.AddProductExpiry(r.Context(), epharma.AddProductExpiryParams{
			ID:         stockin.ProductID,
			ExpiryDate: stockin.ExpiryDate,
		})
	}
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// update product quantity
	err = qtx.IncrementProductQuantity(r.Context(), epharma.IncrementProductQuantityParams{
		ID:       stockin.ProductID,
		Quantity: stockin.Quantity,
	})

	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	tx.Commit(r.Context())
	egor.Redirect(w, r, fmt.Sprintf("/invoices/view/%d", stockin.InvoiceID))
}

func (h *Handlers) DeleteStockIn(w http.ResponseWriter, r *http.Request) {
	stockinID := egor.ParamInt(r, "stockin_id")
	invoiceID := egor.ParamInt(r, "invoice_id")

	tx, err := h.Conn.Begin(r.Context())
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := h.Queries.WithTx(tx)

	// decrement product quantity
	stockin, err := h.Queries.GetStockIn(r.Context(), int32(stockinID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// decrement product quantity
	err = qtx.DecrementProductQuantity(r.Context(), epharma.DecrementProductQuantityParams{
		ID:       stockin.ProductID,
		Quantity: stockin.Quantity,
	})

	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// delete stock in
	err = qtx.DeleteStockIn(r.Context(), int32(stockinID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// remove expiry date from product expiry dates
	err = qtx.RemoveProductExpiry(r.Context(), epharma.RemoveProductExpiryParams{
		ID:         stockin.ProductID,
		ExpiryDate: stockin.ExpiryDate,
	})

	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	// commit the transaction
	tx.Commit(r.Context())

	egor.Redirect(w, r, fmt.Sprintf("/invoices/view/%d", invoiceID))
}
