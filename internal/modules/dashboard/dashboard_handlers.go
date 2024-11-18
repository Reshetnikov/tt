package dashboard

import (
	"log/slog"
	"net/http"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

type DashboardHandlers struct {
	repo *DashboardRepositoryPostgres
}

func NewDashboardHandler(repo *DashboardRepositoryPostgres) *DashboardHandlers {
	return &DashboardHandlers{repo: repo}
}

func (h *DashboardHandlers) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RedirectLogin(w, r)
		return
	}
	tasks := h.repo.Tasks(user.ID, "")
	records := h.repo.RecordsWithTasks(FilterRecords{
		UserID: user.ID,
		// RecordID: 0,
		// Start:    time.Now().Add(-7 * 24 * time.Hour),
		// End:      time.Now(),
	})

	// if r.Header.Get("HX-Request") == "" {
	utils.RenderTemplate(w, []string{"dashboard/dashboard", "dashboard/task_list", "dashboard/record_list"}, utils.TplData{
		"Title":   "Tasks & Records Dashboard",
		"Tasks":   tasks,
		"Records": records,
		"User":    user,
	})
	// } else {
	// 	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/dashboard", "dashboard/task_list", "dashboard/record_list"}, "content", utils.TplData{
	// 		"Title":   "Tasks & Records Dashboard",
	// 		"Tasks":   tasks,
	// 		"Records": records,
	// 	})
	// }

}

var D = slog.Debug
