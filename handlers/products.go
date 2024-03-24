package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/abiiranathan/dbtypes"
	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
)

// RenderProductCreatePage
func (h *Handlers) RenderProductCreatePage(w http.ResponseWriter, r *http.Request) {
	egor.Render(w, r, "products/create.html", egor.Map{})
}

// RenderProductUpdatePage
func (h *Handlers) RenderProductUpdatePage(w http.ResponseWriter, r *http.Request) {
	productID := egor.ParamInt(r, "id")
	product, err := h.Queries.GetProduct(r.Context(), int32(productID))
	if err != nil {
		egor.SendError(w, r, err)
		return
	}

	egor.Render(w, r, "products/update.html", egor.Map{
		"product": product,
	})
}

// RenderProductImportPage
func (h *Handlers) RenderProductImportPage(w http.ResponseWriter, r *http.Request) {
	egor.Render(w, r, "products/import.html", egor.Map{})
}

func parseMoney(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, nil
	}

	money, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid money: %s", s)
	}
	return money, nil
}

func parseQuantity(s string) (int32, error) {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, nil
	}

	quantity, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid quantity: %s", s)
	}
	return int32(quantity), nil
}

// ImportProducts
func (h *Handlers) ImportProducts(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		egor.SendError(w, r, fmt.Errorf("r.FormFile(): error parsing csv file: %s", err), http.StatusBadRequest)
	}
	defer file.Close()

	// parse the csv file
	// create a new product for each row
	// headers:
	// generic_name,brand_name,quantity,expiry_dates,cost_price,selling_price,barcode
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 7
	reader.LazyQuotes = true
	reader.Comma = ','
	reader.Comment = '#'
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		egor.SendError(w, r, fmt.Errorf("reader.ReadAll(): error parsing csv file: %s", err), http.StatusBadRequest)
		return
	}

	// Skip the header
	records = records[1:]
	products := make([]epharma.CreateProductsParams, 0, len(records))
	for _, record := range records {
		var (
			expiryDates  []dbtypes.Date
			err          error
			costPrice    float64
			sellingPrice float64
			quantity     int32
		)

		quantity, err = parseQuantity(record[2])
		if err != nil {
			egor.SendError(w, r, err, http.StatusBadRequest)
			return
		}

		costPrice, err = parseMoney(record[4])
		if err != nil {
			egor.SendError(w, r, err, http.StatusBadRequest)
			return
		}

		sellingPrice, err = parseMoney(record[5])
		if err != nil {
			egor.SendError(w, r, err, http.StatusBadRequest)
			return
		}

		// Parse the expiry dates, "yyyy-mm-dd,yyyy-mm-dd,yyyy-mm-dd"
		exp_dates := strings.Split(record[3], ",")
		for _, date := range exp_dates {
			date = strings.TrimSpace(date)
			if date == "" {
				continue
			}

			expiryDate, err := dbtypes.ParseDateFromString(date)
			if err != nil {
				egor.SendError(w, r, fmt.Errorf("dbtypes.ParseDateFromString(): invalid expiry date: %s", err), http.StatusBadRequest)
				return
			}

			if !expiryDate.IsZero() {
				expiryDates = append(expiryDates, expiryDate)
			}
		}

		product := epharma.CreateProductsParams{
			GenericName:  record[0],
			BrandName:    record[1],
			Quantity:     quantity,
			CostPrice:    costPrice,
			SellingPrice: sellingPrice,
			ExpiryDates:  expiryDates,
			Barcode:      record[6],
		}
		products = append(products, product)
	}

	n, err := h.Queries.CreateProducts(r.Context(), products)
	if err != nil {
		egor.SendError(w, r, fmt.Errorf("h.Queries.CreateProducts(): error creating products: %s", err), http.StatusBadRequest)
		return
	}

	log.Printf("Successfully imported %d products\n", n)
	egor.Redirect(w, r, "/products", http.StatusSeeOther)

}

// CreateProduct
func (h *Handlers) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var params epharma.CreateProductParams
	err := egor.BodyParser(r, &params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	product, err := h.Queries.CreateProduct(r.Context(), params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Redirect(w, r, fmt.Sprintf("/products/view/%d", product.ID), http.StatusSeeOther)
}

// ListProductsPaginated
func (h *Handlers) ListProductsPaginated(w http.ResponseWriter, r *http.Request) {
	page := egor.QueryInt(r, "page", 1)
	limit := egor.QueryInt(r, "limit", 50)
	name := r.URL.Query().Get("name")

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	products, err := h.Queries.ListProductsPaginated(r.Context(), epharma.ListProductsPaginatedParams{
		Off:  int32(offset),
		Lim:  int32(limit),
		Name: name,
	})

	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	count, err := h.Queries.CountProducts(r.Context())
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	var totalPages int64
	if count%int64(limit) == 0 {
		totalPages = count / int64(limit)
	} else {
		totalPages = count/int64(limit) + 1
	}

	egor.Render(w, r, "products/list.html", egor.Map{
		"products":   products,
		"PageSize":   limit,
		"Count":      count,
		"Page":       page,
		"TotalPages": totalPages,
		"HasNext":    int64(page) < totalPages,
		"HasPrev":    page > 1,
	})
}

// GetProduct
func (h *Handlers) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := egor.ParamInt(r, "id")
	product, err := h.Queries.GetProduct(r.Context(), int32(productID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "products/view.html", egor.Map{
		"product": product,
	})
}

// UpdateProduct
func (h *Handlers) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID := egor.ParamInt(r, "id")

	var params epharma.UpdateProductParams
	err := egor.BodyParser(r, &params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	params.ID = int32(productID)

	err = h.Queries.UpdateProduct(r.Context(), params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Redirect(w, r, "/products", http.StatusSeeOther)
}

// DeleteProduct
func (h *Handlers) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID := egor.ParamInt(r, "id")
	err := h.Queries.DeleteProduct(r.Context(), int32(productID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Redirect(w, r, "/products", http.StatusSeeOther)
}

// GetProductByBarcode
func (h *Handlers) GetProductByBarcode(w http.ResponseWriter, r *http.Request) {
	barcode := r.PathValue("barcode")
	queryType := egor.Query(r, "type", "json")
	if queryType != "json" && queryType != "html" {
		panic("invalid query type")
	}

	product, err := h.Queries.GetProductByBarcode(r.Context(), barcode)
	if err != nil {
		if queryType == "json" {
			egor.SendJSONError(w, map[string]any{"error": err.Error()}, http.StatusBadRequest)
		} else {
			egor.SendError(w, r, err)
		}
		return
	}

	if queryType == "json" {
		egor.SendJSON(w, product)
	} else {
		buf := new(bytes.Buffer)
		err := egor.ExecuteTemplate(buf, r, "products/results.html", egor.Map{
			"products": []epharma.Product{product},
		})

		if err != nil {
			egor.SendError(w, r, err)
			return
		}

		egor.SendHTML(w, buf.String())
	}
}

// SearchProducts
func (h *Handlers) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := egor.Query(r, "name")
	limit := egor.QueryInt(r, "limit", 5)
	retType := egor.Query(r, "type", "json")

	if retType != "json" && retType != "html" {
		panic("invalid return type")
	}

	products, err := h.Queries.SearchProducts(r.Context(), query)
	if err != nil {
		egor.SendJSONError(w, map[string]any{"error": err.Error()}, http.StatusBadRequest)
		return
	}

	if len(products) > limit {
		products = products[:limit]
	}

	if retType == "json" {
		egor.SendJSON(w, products)
	} else {
		// execute view search_results.html with .products
		buf := new(bytes.Buffer)
		err := egor.ExecuteTemplate(buf, r, "products/results", egor.Map{
			"products": products,
		})

		w.Header().Set("Content-Type", "text/html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error: %s", err.Error())
			return
		}

		w.Write(buf.Bytes())
	}
}
