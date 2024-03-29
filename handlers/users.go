package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/abiiranathan/egor/egor"
	"github.com/abiiranathan/epharmacy/epharma"
)

// CreateUser
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.Queries.CreateUser(r.Context(), epharma.CreateUserParams{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	})

	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Redirect(w, r, fmt.Sprintf("/users/%d", user.ID))
}

// ListUsers
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Queries.ListUsers(r.Context())
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Render(w, r, "accounts/list", egor.Map{
		"users": users,
		"breadcrumbs": Breadcrumbs{
			{Label: "Users", IsLast: true},
		},
	})
}

func (h *Handlers) RenderUserCreatePage(w http.ResponseWriter, r *http.Request) {
	egor.Render(w, r, "accounts/create", egor.Map{
		"breadcrumbs": Breadcrumbs{
			{Label: "Users", URL: "/users"},
			{Label: "Create User", IsLast: true},
		},
	})
}

func (h *Handlers) RenderUserEditPage(w http.ResponseWriter, r *http.Request) {
	userId := egor.ParamInt(r, "id")
	user, err := h.Queries.GetUser(r.Context(), int32(userId))
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	egor.Render(w, r, "accounts/update", egor.Map{
		"user": user,
		"breadcrumbs": Breadcrumbs{
			{Label: "Users", URL: "/users"},
			{Label: user.Username, URL: fmt.Sprintf("/users/%d", user.ID)},
			{Label: "Update User", IsLast: true},
		},
	})
}

// GetUser
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	user, err := h.Queries.GetUser(r.Context(), int32(userID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusNotFound)
		return
	}

	egor.Render(w, r, "accounts/view", egor.Map{
		"user": user,
		"breadcrumbs": Breadcrumbs{
			{Label: "Users", URL: "/users"},
			{Label: user.Username, IsLast: true},
		},
	})
}

// UpdateUser
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	var params epharma.UpdateUserParams
	err := egor.BodyParser(r, &params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}

	params.ID = int32(userID)
	params.UpdatePassword = params.Password != ""

	err = h.Queries.UpdateUser(r.Context(), params)
	if err != nil {
		egor.SendError(w, r, err, http.StatusBadRequest)
		return
	}
	egor.Redirect(w, r, fmt.Sprintf("/users/%d", userID))
}

// DeleteUser
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	err := h.Queries.DeleteUser(r.Context(), int32(userID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusInternalServerError)
		return
	}
	egor.Redirect(w, r, "/users")
}

// ActivateUser
func (h *Handlers) ActivateUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	err := h.Queries.ActivateUser(r.Context(), int32(userID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusInternalServerError)
		return
	}
	egor.Redirect(w, r, "/users")
}

// DeActivateUser
func (h *Handlers) DeActivateUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	err := h.Queries.DeactivateUser(r.Context(), int32(userID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusInternalServerError)
		return
	}
	egor.Redirect(w, r, "/users")
}

// PromoteUser
func (h *Handlers) PromoteUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	err := h.Queries.PromoteUser(r.Context(), int32(userID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusInternalServerError)
		return
	}
	egor.Redirect(w, r, "/users")
}

// DemoteUser
func (h *Handlers) DemoteUser(w http.ResponseWriter, r *http.Request) {
	userID := egor.ParamInt(r, "id")
	err := h.Queries.DemoteUser(r.Context(), int32(userID))
	if err != nil {
		egor.SendError(w, r, err, http.StatusInternalServerError)
		return
	}
	egor.Redirect(w, r, "/users")
}

// Auth
// RenderLoginPage
func (h *Handlers) RenderLoginPage(w http.ResponseWriter, r *http.Request) {
	egor.Render(w, r, "login", egor.Map{
		"breadcrumbs": Breadcrumbs{
			{Label: "Login", IsLast: true},
		},
	})
}

func encodeUserId(id int32) string {
	// base64 encode the user id using 32 bytes
	userIdStr := fmt.Sprintf("%d", id)

	return base64.StdEncoding.EncodeToString([]byte(userIdStr))
}

func decodeUserId(encodedId string) (int32, error) {
	// base64 decode the user id
	userIdStr, err := base64.StdEncoding.DecodeString(encodedId)
	if err != nil {
		return 0, err
	}

	userId, err := strconv.Atoi(string(userIdStr))
	if err != nil {
		return 0, err
	}

	return int32(userId), nil
}

// Login
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var invalidCtx = egor.Map{
		"error":    "Invalid username or password",
		"username": username,
		"password": password,
		"breadcrumbs": Breadcrumbs{
			{Label: "Login", URL: "/login", IsLast: true},
		},
	}

	user, err := h.Queries.GetUserByUsername(r.Context(), username)
	if err != nil {
		egor.Render(w, r, "login", invalidCtx)
		return
	}

	if user.Password != password {
		egor.Render(w, r, "login", invalidCtx)
		return
	}

	// Set the cookie to the request
	cookie := http.Cookie{
		Name:     "user_id",
		Value:    encodeUserId(user.ID),
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
	}

	http.SetCookie(w, &cookie)

	next := r.URL.Query().Get("next")
	if next == "" {
		next = "/"
	}

	egor.Redirect(w, r, next)
}

// Logout
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// get the cookie and make it expire
	cookie := http.Cookie{
		Name:   "user_id",
		Path:   "/",
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
	egor.Redirect(w, r, "/login")
}
