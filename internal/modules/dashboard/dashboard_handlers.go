package dashboard

import (
	"log/slog"
	"net/http"
	"time-tracker/internal/modules/users"
)

type DashboardHandler struct {
	service *DashboardService
}

func NewDashboardHandler(service *DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		slog.Debug("dashboard redirect")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	slog.Info("dashboard", "user", user)

	// data, err := h.service.GetDashboardData(user.ID)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// tmpl, err := template.ParseFiles("./web/templates/dashboard.html")
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// tmpl.Execute(w, data)
}
