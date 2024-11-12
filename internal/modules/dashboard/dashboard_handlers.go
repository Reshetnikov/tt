package dashboard

import (
	"log/slog"
	"net/http"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

var D = slog.Debug

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
	records, tasks := h.repo.RecordsAndTasks(user.ID)

	D("tasks", "tasks", tasks)
	D("HandleDashboard", "records", records)

	if r.Header.Get("HX-Request") == "" {
		users.RenderTemplate(w, r, []string{"dashboard/dashboard", "dashboard/task_list", "dashboard/record_list"}, utils.TplData{
			"Title":   "Tasks & Records Dashboard",
			"Tasks":   tasks,
			"Records": records,
		})
	} else {
		utils.RenderTemplateWithoutLayout(w, []string{"dashboard/dashboard", "dashboard/task_list", "dashboard/record_list"}, "content", utils.TplData{
			"Title":   "Tasks & Records Dashboard",
			"Tasks":   tasks,
			"Records": records,
		})
	}

}

func (h *DashboardHandler) HandleRecordList(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}
	records, _ := h.repo.RecordsAndTasks(user.ID)
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/record_list"}, "dashboard/record_list", utils.TplData{
		"Records": records,
	})
}
