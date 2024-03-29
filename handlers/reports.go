package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/abiiranathan/dbtypes"
	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
)

type WeelyIncome struct {
	WeekDays []string
	Income   []float64
}

type MonthlyIncome struct {
	Months []string
	Income []float64
}

type AnnualIncome struct {
	Years  []int
	Income []float64
}

func (h *Handlers) fetchDashboard(ctx context.Context) (
	dailyProductSales []epharma.SalesReport,
	monthlyProductSales []epharma.MonthlySalesReportsRow,
	annualProductSales []epharma.AnnualSalesReportsRow,
	err error,
) {

	dailyProductSales, err = h.Queries.DailySalesReports(ctx, "")
	if err != nil {
		return
	}

	monthlyProductSales, err = h.Queries.MonthlySalesReports(ctx, "")
	if err != nil {
		return
	}

	annualProductSales, err = h.Queries.AnnualSalesReports(ctx, "")
	if err != nil {
		return
	}

	return
}

func dateIsToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day()
}

func dateIsThisWeek(date time.Time) bool {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 6)
	return date.After(weekStart) && date.Before(weekEnd)
}

func dateIsThisMonth(date time.Time) bool {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)
	return date.After(monthStart) && date.Before(monthEnd)
}

func dateIsThisYear(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year()
}

func (h *Handlers) RenderReportsDashboard(w http.ResponseWriter, r *http.Request) {
	dailyProductSales, monthlyProductSales, annualProductSales, err := h.fetchDashboard(r.Context())
	if err != nil {
		egor.SendError(w, r, err)
		log.Printf("Error fetching data for dashboard: %v\n", err)
		return
	}

	// Aggregate data for today, this week, this month, this year
	var incomeToday, incomeThisWeek, incomeThisMonth, incomeThisYear float64

	for _, sale := range dailyProductSales {
		if dateIsToday(time.Time(sale.TransactionDate)) {
			incomeToday += float64(sale.TotalIncome)
		}

		if dateIsThisWeek(time.Time(sale.TransactionDate)) {
			incomeThisWeek += float64(sale.TotalIncome)
		}

		if dateIsThisMonth(time.Time(sale.TransactionDate)) {
			incomeThisMonth += float64(sale.TotalIncome)
		}

		if dateIsThisYear(time.Time(sale.TransactionDate)) {
			incomeThisYear += float64(sale.TotalIncome)
		}
	}

	// create a slice of weekly income
	weeklyIncome := WeelyIncome{
		WeekDays: []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		Income:   make([]float64, 7),
	}

	for _, sale := range dailyProductSales {
		if dateIsThisWeek(time.Time(sale.TransactionDate)) {
			weeklyIncome.Income[int(time.Time(sale.TransactionDate).Weekday())] += float64(sale.TotalIncome)
		}
	}

	// create a slice of monthly income
	monthlyIncome := MonthlyIncome{
		Months: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		Income: make([]float64, 12),
	}

	for _, sale := range monthlyProductSales {
		if sale.Month.Year() == time.Now().Year() {
			monthlyIncome.Income[sale.Month.Month()-1] += float64(sale.TotalIncome)
		}
	}

	// annual income
	annualIncome := AnnualIncome{
		Years:  []int{time.Now().Year() - 2, time.Now().Year() - 1, time.Now().Year()},
		Income: make([]float64, 3),
	}

	for _, sale := range annualProductSales {
		if sale.Year.Year() == time.Now().Year() {
			annualIncome.Income[2] += float64(sale.TotalIncome)
		} else if sale.Year.Year() == time.Now().Year()-1 {
			annualIncome.Income[1] += float64(sale.TotalIncome)
		} else if sale.Year.Year() == time.Now().Year()-2 {
			annualIncome.Income[0] += float64(sale.TotalIncome)
		}
	}

	// Convert the weekly and monthly income to JSON
	weeklyIncomeJSON, err := json.Marshal(weeklyIncome)
	if err != nil {
		egor.SendError(w, r, err)
		log.Printf("Error converting weekly income to JSON: %v\n", err)
		return
	}

	monthlyIncomeJSON, err := json.Marshal(monthlyIncome)
	if err != nil {
		egor.SendError(w, r, err)
		log.Printf("Error converting monthly income to JSON: %v\n", err)
		return
	}

	annualIncomeJSON, err := json.Marshal(annualIncome)
	if err != nil {
		egor.SendError(w, r, err)
		log.Printf("Error converting annual income to JSON: %v\n", err)
		return
	}

	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 6)

	const dailyLimit int = 14
	if len(dailyProductSales) > dailyLimit {
		dailyProductSales = dailyProductSales[:dailyLimit]
	}

	egor.Render(w, r, "reports/dashboard.html", egor.Map{
		"Today":     time.Now(),
		"WeekStart": weekStart,
		"WeekEnd":   weekEnd,

		"dailyProductSales":   dailyProductSales,
		"monthlyProductSales": monthlyProductSales,
		"annualProductSales":  annualProductSales,

		"incomeToday":     incomeToday,
		"incomeThisWeek":  incomeThisWeek,
		"incomeThisMonth": incomeThisMonth,
		"incomeThisYear":  incomeThisYear,

		"WeeklyIncome":  string(weeklyIncomeJSON),
		"MonthlyIncome": string(monthlyIncomeJSON),
		"AnnualIncome":  string(annualIncomeJSON),
		"breadcrumbs": Breadcrumbs{
			{Label: "Dashboard", URL: "/reports", IsLast: true},
		},
	})
}

func (h *Handlers) DailyProductSalesReport(w http.ResponseWriter, r *http.Request) {
	date := egor.Query(r, "date") // Format: "yyyy-mm-dd"
	if date == "" {
		egor.SendError(w, r, fmt.Errorf("missing date query parameter"), http.StatusBadRequest)
		return
	}

	dailyProductSales, err := h.Queries.DailyProductSales(r.Context(), date)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	dateObj, _ := dbtypes.ParseDateFromString(date)
	egor.Render(w, r, "reports/daily_product_sales.html", egor.Map{
		"DailyProductSales": dailyProductSales,
		"Date":              dateObj,
		"breadcrumbs": Breadcrumbs{
			{Label: "Dashboard", URL: "/reports"},
			{Label: "Daily Product Sales", IsLast: true},
		},
	})
}

func (h *Handlers) MonthlyProductSalesReport(w http.ResponseWriter, r *http.Request) {
	date := egor.Query(r, "month") // format: "12-2024"
	if date == "" {
		egor.SendError(w, r, fmt.Errorf("missing date query parameter"), http.StatusBadRequest)
		return
	}

	// split the date into month and year
	parts := strings.SplitN(date, "-", 2)
	if len(parts) != 2 {
		egor.SendError(w, r, fmt.Errorf("invalid date format"), http.StatusBadRequest)
		return
	}

	// create a truncated date
	date = fmt.Sprintf("%s-%s-01", parts[1], parts[0])
	fmt.Printf("Date: %s\n", date)

	monthlyProductSales, err := h.Queries.MonthlyProductSales(r.Context(), date)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	dateObj, _ := dbtypes.ParseDateFromString(date)

	egor.Render(w, r, "reports/monthly_product_sales.html", egor.Map{
		"MonthlyProductSales": monthlyProductSales,
		"Date":                dateObj,
		"breadcrumbs": Breadcrumbs{
			{Label: "Dashboard", URL: "/reports"},
			{Label: "Monthly Product Sales", IsLast: true},
		},
	})
}

func (h *Handlers) AnnualProductSalesReport(w http.ResponseWriter, r *http.Request) {
	date := egor.Query(r, "year") // format: "2024"
	if date == "" {
		egor.SendError(w, r, fmt.Errorf("missing date query parameter"), http.StatusBadRequest)
		return
	}

	// create a truncated date
	date = fmt.Sprintf("%s-01-01", date)

	annualProductSales, err := h.Queries.AnnualProductSales(r.Context(), date)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	dateObj, _ := dbtypes.ParseDateFromString(date)

	egor.Render(w, r, "reports/annual_product_sales.html", egor.Map{
		"AnnualProductSales": annualProductSales,
		"Date":               dateObj,
		"breadcrumbs": Breadcrumbs{
			{Label: "Dashboard", URL: "/reports"},
			{Label: "Annual Product Sales", IsLast: true},
		},
	})
}
