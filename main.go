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

func loadTemplates() *template.Template {
	tmpl, err := egor.ParseTemplatesRecursiveFS(views, "views", handlers.FuncMap, ".html")
	if err != nil {
		panic(err)
	}
	return tmpl
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("EPHARMA_PG_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	// create a router and apply the middlewares
	options := []egor.RouterOption{
		egor.WithTemplates(loadTemplates()),
		egor.BaseLayout("base.html"),
		egor.ContentBlock("content"),
		egor.ErrorTemplate("error.html"),
		egor.PassContextToViews(true),
	}

	router := egor.NewRouter(options...)
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
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	addr := fmt.Sprintf(":%s", port)

	// create a new server
	server := egor.NewServer(addr, router)
	defer server.GracefulShutdown(time.Second * 10)

	// Run the server forever
	fmt.Printf("Listening on port %s\n", port)
	log.Fatalln(server.ListenAndServe())
}
