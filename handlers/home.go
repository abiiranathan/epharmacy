package handlers

import (
	"net/http"

	"github.com/abiiranathan/egor/egor"
)

// Home page render sales table and the most common products..
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	limit := egor.ParamInt(r, "limit", 5)
	products, err := h.Queries.MostCommonProducts(r.Context(), int32(limit))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "index.html", egor.Map{
		"products": products,
	})
}
