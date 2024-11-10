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
	records, tasks := h.repo.RecordsWithTasks(user.ID)

	D("tasks", "tasks", tasks)
	D("HandleDashboard", "records", records)

	users.RenderTemplate(w, r, "dashboard/dashboard", utils.TplData{
		"Title":   "Tasks & Records Dashboard",
		"Tasks":   tasks,
		"Records": records,
	})

	// utils.RenderTemplateWithoutLayout(w, "dashboard/dashboard", "content", utils.TplData{
	// 	"Title":   "Tasks & Records Dashboard",
	// 	"Tasks":   tasks,
	// 	"Records": records,
	// })
}

type taskForm struct {
	Title       string `form:"title" validate:"required"`
	Description string `form:"description" validate:"required"`
	Color       string `form:"color" validate:"required"`
}

func (h *DashboardHandler) HandleTasksNew(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplateWithoutLayout(w, "dashboard/form_task", "dashboard/form_task", utils.TplData{
		"Errors": utils.FormErrors{},
		"Form":   taskForm{},
	})
}
