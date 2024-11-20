package dashboard

import (
	"log/slog"
	"net/http"
	"time"
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

	week := r.URL.Query().Get("week")
	nowWithTimezone, _ := utils.NowWithTimezone(user.TimeZone)
	startInterval, endInterval := GetDateInterval(week, nowWithTimezone)

	filterRecords := FilterRecords{
		UserID:        user.ID,
		StartInterval: startInterval,
		EndInterval:   endInterval,
	}

	dailyRecords := h.repo.DailyRecords(filterRecords, nowWithTimezone)

	tasks := h.repo.Tasks(user.ID, "")

	previousWeek := utils.FormatISOWeek(startInterval.AddDate(0, 0, -7))
	nextWeek := utils.FormatISOWeek(endInterval.AddDate(0, 0, 7))
	utils.RenderTemplate(w, []string{"dashboard/dashboard", "dashboard/task_list", "dashboard/record_list"}, utils.TplData{
		"Title":        "Tasks & Records Dashboard",
		"Tasks":        tasks,
		"DailyRecords": dailyRecords,
		"User":         user,
		"Week":         utils.FormatISOWeek(startInterval),
		"PreviousWeek": previousWeek,
		"NextWeek":     nextWeek,
	})

}

func GetDateInterval(week string, nowWithTimezone time.Time) (startInterval time.Time, endInterval time.Time) {
	if week != "" {
		var err error
		startInterval, endInterval, err = utils.GetWeekInterval(week)
		if err == nil {
			return
		}
	}
	startInterval, endInterval = utils.GetDateInterval(nowWithTimezone)
	return
}

var D = slog.Debug
