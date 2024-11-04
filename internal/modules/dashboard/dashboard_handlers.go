package dashboard

import (
	"net/http"
	"strconv"

	"html/template"
)

type DashboardHandler struct {
	service *DashboardService
}

func NewDashboardHandler(service *DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Context().Value("userID").(string)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	data, err := h.service.GetDashboardData(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./web/templates/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}