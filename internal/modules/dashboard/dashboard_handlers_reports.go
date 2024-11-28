package dashboard

import (
	"net/http"
	"time"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

func (h *DashboardHandlers) HandleReports(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RedirectLogin(w, r)
		return
	}

	monthStr := r.URL.Query().Get("month")
	nowWithTimezone, _ := utils.NowWithTimezone(user.TimeZone)

	startInterval, endInterval := getMonthInterval(monthStr, nowWithTimezone)

	filterRecords := FilterRecords{
		UserID:        user.ID,
		StartInterval: startInterval,
		EndInterval:   endInterval,
	}

	//dailyRecords := h.repo.DailyRecords(filterRecords, nowWithTimezone)
	D("HandleReports", "filterRecords", filterRecords)
	//D("HandleReports", "dailyRecords", dailyRecords)

	utils.RenderTemplate(w, []string{"dashboard/reports"}, utils.TplData{
		"Title": "Reports",
		"User":  user,
	})
}

func getMonthInterval(monthStr string, nowWithTimezone time.Time) (startInterval time.Time, endInterval time.Time) {
	if monthStr != "" {
		parsedTime, err := time.Parse("2006-01", monthStr)
		if err == nil {
			startInterval = parsedTime
		}
	}
	if startInterval.IsZero() {
		year, month, _ := nowWithTimezone.Date()
		startInterval = time.Date(year, month, 1, 0, 0, 0, 0, nowWithTimezone.Location())
	}
	endInterval = startInterval.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return
}
