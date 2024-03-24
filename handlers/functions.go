package handlers

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/abiiranathan/dbtypes"
	"github.com/abiiranathan/epharmacy/epharma"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const warnAtMonths = 2

var FuncMap = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("02 Jan 2006")
	},
	"formatDateTime": func(t time.Time) string {
		return t.Format("02 Jan 2006 15:04")
	},
	"multiply": func(a, b int) int {
		return a * b
	},
	"product_subtotal": func(product epharma.Product) float64 {
		return product.SellingPrice * float64(product.Quantity)
	},
	"transaction_total": func(transaction Transaction) float64 {
		var total float64
		for _, product := range transaction.Products {
			total += product.SellingPrice * float64(product.Quantity)
		}
		return total
	},
	"invoice_subtotal": func(invoice epharma.InvoiceItemsRow) float64 {
		return invoice.CostPrice * float64(invoice.Quantity)
	},
	"roundf64": func(f float64) string {
		return CurrencyF64(f)
	},
	"minus": func(a, b int) int {
		return a - b
	},
	"minusf": func(a, b float64) float64 {
		return a - b
	},
	"plus": func(a, b int) int {
		return a + b
	},
	"join_dates": func(arr []dbtypes.Date) string {
		var dates []string
		for _, date := range arr {
			dates = append(dates, date.String())
		}
		return strings.Join(dates, ", ")
	},
	"days_to_expiry": daysToExpiry,
	"expiryColor": func(expiry dbtypes.Date) string {
		if expiry.IsZero() {
			return "text-gray-700"
		}

		n := int(time.Until(time.Time(expiry)).Hours() / 24)

		if n < 0 {
			return "text-red-700 bg-red-100 p-1 rounded mb-1"
		} else if n < warnAtMonths*30 {
			return "text-yellow-600 bg-yellow-100 p-1 rounded mb-1"
		} else {
			return "text-green-700 bg-green-100 p-1 rounded mb-1"
		}
	},
	"CurrencyF64": CurrencyF64,
}

func formatDuration(duration time.Duration) string {
	n := int(duration.Hours() / 24)

	var years, months, days int
	if n > 365 {
		years = n / 365
		months = (n % 365) / 30
		return fmt.Sprintf("%d yrs, %d months", years, months)
	} else if n > 30 {
		months = n / 30
		days = n % 30
		if days >= 14 {
			return fmt.Sprintf("%d months, %d days", months, days)
		} else {
			return fmt.Sprintf("%d months", months)
		}
	} else {
		return fmt.Sprintf("%d days", n)
	}
}

func daysToExpiry(expiry dbtypes.Date) string {
	if expiry.IsZero() {
		return "N/A"
	}

	duration := time.Until(time.Time(expiry))
	if duration >= 0 {
		return formatDuration(duration)
	}

	duration = duration * -1 // Make duration positive
	return fmt.Sprintf("EXP: %s ago", formatDuration(duration))
}

func CurrencyF64(number float64) string {
	p := message.NewPrinter(language.BritishEnglish)
	return p.Sprintf("%.2f", number)
}

// func CurrencyInt(number int) string {
// 	p := message.NewPrinter(language.BritishEnglish)
// 	return p.Sprintf("%d", number)
// }

// func CurrencyI64(number int64) string {
// 	p := message.NewPrinter(language.BritishEnglish)
// 	return p.Sprintf("%d", number)
// }
