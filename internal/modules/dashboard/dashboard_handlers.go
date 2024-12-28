package dashboard

import (
	"fmt"
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

	weekStr := r.URL.Query().Get("week")
	nowWithTimezone, _ := utils.NowWithTimezone(user.TimeZone)
	startInterval, endInterval := getWeekInterval(weekStr, nowWithTimezone, user.IsWeekStartMonday)

	filterRecords := FilterRecords{
		UserID:        user.ID,
		StartInterval: startInterval,
		EndInterval:   endInterval,
	}

	dailyRecords := h.repo.DailyRecords(filterRecords, nowWithTimezone)

	tasks := h.repo.Tasks(user.ID, "")

	previousWeek := utils.FormatISOWeek(startInterval.AddDate(0, 0, -7), user.IsWeekStartMonday)
	nextWeek := utils.FormatISOWeek(endInterval.AddDate(0, 0, 7), user.IsWeekStartMonday)
	utils.RenderTemplate(w, []string{"dashboard/dashboard", "dashboard/task_list", "dashboard/record_list", "dashboard/record_list_navigation"}, utils.TplData{
		"Title":           "Tasks & Records Dashboard",
		"Tasks":           tasks,
		"DailyRecords":    dailyRecords,
		"User":            user,
		"Week":            utils.FormatISOWeek(startInterval, user.IsWeekStartMonday),
		"PreviousWeek":    previousWeek,
		"NextWeek":        nextWeek,
		"NowWithTimezone": nowWithTimezone,
	})

}

func getWeekInterval(weekStr string, nowWithTimezone time.Time, isWeekStartMonday bool) (startInterval time.Time, endInterval time.Time) {
	if weekStr != "" {
		var err error
		startInterval, endInterval, err = utils.GetWeekInterval(weekStr, isWeekStartMonday)
		if err == nil {
			return
		}
	}
	startInterval, endInterval = utils.GetWeekIntervalByDate(nowWithTimezone, isWeekStartMonday)
	return
}

var D = slog.Debug
var P = fmt.Println
