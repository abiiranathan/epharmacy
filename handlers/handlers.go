package handlers

import (
	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
	"github.com/jackc/pgx/v5"
)

type Handlers struct {
	Queries *epharma.Queries
	Router  *egor.Router
	Conn    *pgx.Conn
}

func New(queries *epharma.Queries, conn *pgx.Conn, router *egor.Router) *Handlers {
	return &Handlers{
		Queries: queries,
		Router:  router,
		Conn:    conn,
	}
}
