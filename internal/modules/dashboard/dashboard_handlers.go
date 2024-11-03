package dashboard

import (
	"net/http"
	"strconv"

	"html/template"
)

type DashboardHandler struct {
	service *DashboardService
}

// NewDashboardHandler создает новый обработчик для панели управления
func NewDashboardHandler(service *DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// HandleDashboard обрабатывает запросы к панели управления
func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (например, из сессии)
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