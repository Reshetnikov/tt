package dashboard

import (
	"log/slog"
	"net/http"
	"time"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

var d = slog.Debug

type DashboardHandler struct {
	repo *DashboardRepositoryPostgres
}

func NewDashboardHandler(repo *DashboardRepositoryPostgres) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RedirectLogin(w, r)
		return
	}
	tasks := h.repo.FetchTasks(user.ID)
	selectedWeek := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday())) // Начало недели
	weeklyRecords := h.repo.FetchWeeklyRecords(user.ID, selectedWeek)

	// d("HandleDashboard", "tasks", tasks)
	d("HandleDashboard", "weeklyRecords", weeklyRecords)

	users.RenderTemplate(w, r, "dashboard", utils.TplData{
		"Title":         "Tasks & Records Dashboard",
		"Tasks":         tasks,
		"WeeklyRecords": weeklyRecords,
		"SelectedWeek":  selectedWeek,
	})
}
