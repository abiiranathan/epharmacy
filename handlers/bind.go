package handlers

func ConnectRoutes(h *Handlers) {
	// auth
	h.Router.Get("/login", h.RenderLoginPage)
	h.Router.Post("/login", h.Login)
	h.Router.Post("/logout", h.Logout)

	// Auth middleware
	h.Router.Use(h.AuthMiddleware)

	// Home page
	h.Router.Get("/", h.Home)

	// Users
	users := h.Router.Group("/users", h.AdminRequired)
	users.Post("/", h.CreateUser)
	users.Get("/", h.ListUsers)
	users.Get("/{id}", h.GetUser)
	users.Post("/{id}", h.UpdateUser)
	users.Post("/delete/{id}", h.DeleteUser)
	users.Get("/new/account", h.RenderUserCreatePage)
	users.Get("/edit/{id}", h.RenderUserEditPage)

	users.Post("/activate/{id}", h.ActivateUser)
	users.Post("/deactivate/{id}", h.DeActivateUser)
	users.Post("/promote/{id}", h.PromoteUser)
	users.Post("/demote/{id}", h.DemoteUser)

	// Products
	products := h.Router.Group("/products")
	products.Get("/", h.ListProductsPaginated)
	products.Get("/create", h.RenderProductCreatePage)
	products.Post("/create", h.CreateProduct)
	products.Get("/view/{id}", h.GetProduct)
	products.Get("/search", h.SearchProducts)
	products.Get("/search/barcode/{barcode}", h.GetProductByBarcode)
	products.Get("/update/{id}", h.RenderProductUpdatePage)
	products.Post("/update/{id}", h.UpdateProduct)
	products.Post("/delete/{id}", h.DeleteProduct)
	products.Get("/import", h.RenderProductImportPage)
	products.Post("/import", h.ImportProducts)

	// Transactions
	transactions := h.Router.Group("/transactions")
	transactions.Post("/", h.CreateTransaction)
	transactions.Get("/", h.ListTransactionsPaginated)
	transactions.Get("/{id}", h.GetTransaction)
	transactions.Post("/delete/{id}", h.DeleteTransaction)

	// Invoices
	invoices := h.Router.Group("/invoices")
	invoices.Get("/", h.ListInvoicesPaginated)
	invoices.Get("/create", h.RenderInvoiceCreatePage)
	invoices.Post("/create", h.CreateInvoice)
	invoices.Get("/update/{id}", h.RenderInvoiceUpdatePage)
	invoices.Post("/update/{id}", h.UpdateInvoice)
	invoices.Get("/view/{id}", h.GetInvoice)
	invoices.Post("/delete/{id}", h.DeleteInvoice)
	invoices.Get("/search", h.GetInvoiceByNumber)
	invoices.Get("/list-products", h.ListInvoiceProducts)

	// Stock in
	stockin := h.Router.Group("/stockin")
	stockin.Post("/create", h.NewStockIn)
	stockin.Post("/delete/{invoice_id}/{stockin_id}", h.DeleteStockIn)
}
