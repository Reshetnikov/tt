package utils

import (
	"fmt"
	"net/http"
	"net/url"
)

func RedirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func RedirectLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
func RedirectDashboard(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func RedirectSignupSuccess(w http.ResponseWriter, r *http.Request, email string) {
	url := fmt.Sprintf("/signup-success?email=%s", url.QueryEscape(email))
	http.Redirect(w, r, url, http.StatusSeeOther)
}
