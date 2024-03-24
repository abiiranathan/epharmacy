package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
)

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		var id int32
		var userId string
		var err error
		var user epharma.User

		userIdCookier, err := r.Cookie("user_id")
		if err != nil {
			goto loginFailed
		}

		userId = userIdCookier.Value
		if userId == "" {
			goto loginFailed
		}

		id, err = decodeUserId(userId)
		if err != nil {
			goto loginFailed
		}

		user, err = h.Queries.GetUser(r.Context(), id)
		if err != nil {
			goto loginFailed
		}

		if !user.IsActive {
			goto loginFailed
		}

		egor.SetContextValue(r, "user", user)
		next.ServeHTTP(w, r)
		return

	loginFailed:
		nextUrl := r.URL.Path
		if r.URL.RawQuery != "" {
			nextUrl += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, "/login?next="+nextUrl, http.StatusSeeOther)
		if err != nil {
			log.Println(err)
		}
	})
}

func (h *Handlers) AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := egor.GetContextValue(r, "user").(epharma.User)
		if !ok {
			h.Router.RenderError(w, fmt.Errorf("you are not authenticated"), http.StatusForbidden)
			return
		}

		if !user.IsAdmin {
			h.Router.RenderError(w, fmt.Errorf("admin permission is required"), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
