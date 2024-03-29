package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/egor/egor/middleware/cors"

	"github.com/abiiranathan/egor/egor/middleware/etag"
	"github.com/abiiranathan/egor/egor/middleware/logger"
	"github.com/abiiranathan/egor/egor/middleware/recovery"
	"github.com/abiiranathan/epharmacy/epharma"
	"github.com/abiiranathan/epharmacy/handlers"
)

//go:embed all:static
var static embed.FS

//go:embed all:views
var views embed.FS

var port = "8080"

const (
	// Variable to configure template engine.
	TemplateDir   = "views"
	TemplateExt   = ".html"
	BaseTemplate  = "base.html"
	ErrorTemplate = "error.html"
	ContentBlock  = "content"

	// Database connection URI environment variable.
	EPHARMA_PG_URL_ENV = "EPHARMA_PG_URL"

	// Port environment variable.
	PORT_ENV = "PORT"
)

// Loads all templates in TemplateDir recursively.
func loadTemplates() *template.Template {
	tmpl, err := egor.ParseTemplatesRecursiveFS(views, TemplateDir, handlers.FuncMap, TemplateExt)
	if err != nil {
		panic(err)
	}
	return tmpl
}

func main() {
	ctx := context.Background()

	// Connect to database using the pgx/v5 driver.
	conn, err := pgx.Connect(ctx, os.Getenv(EPHARMA_PG_URL_ENV))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	// Slice of router options.
	options := []egor.RouterOption{
		egor.WithTemplates(loadTemplates()),
		egor.BaseLayout(BaseTemplate),
		egor.ContentBlock(ContentBlock),
		egor.ErrorTemplate(ErrorTemplate),
		egor.PassContextToViews(true),
	}

	// create a router and apply the middlewares
	router := egor.NewRouter(options...)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			egor.SetContextValue(r, "Path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	router.Use(recovery.New(true))
	router.Use(logger.New(os.Stderr).Logger)
	router.Use(etag.New())
	router.Use(cors.New())

	// create a new instance of the epharma queries
	queries := epharma.New(conn)
	handler := handlers.New(queries, conn, router)

	// Serve the static files
	router.StaticFS("/static", http.FS(static))
	handlers.ConnectRoutes(handler)

	// Update the port if it's set
	if os.Getenv(PORT_ENV) != "" {
		port = os.Getenv(PORT_ENV)
	}

	addr := fmt.Sprintf(":%s", port)

	// create a new server
	server := egor.NewServer(addr, router)
	defer server.GracefulShutdown(time.Second * 10)

	fmt.Printf("Listening on http://0.0.0.0:%s\n", port)

	// Run the server forever
	log.Fatalln(server.ListenAndServe())
}
