package dashboard

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

type recordForm struct {
	ID        int
	TaskID    int    `form:"task_id" validate:"required"`
	TimeStart string `form:"time_start" validate:"required,datetime=2006-01-02T15:04" label:"Time Start"`
	TimeEnd   string `form:"time_end" validate:"omitempty,datetime=2006-01-02T15:04" label:"Time End"`
	Comment   string `form:"comment" validate:"max=10000"`
}

type recordFormData struct {
	Errors utils.FormErrors
	Form   recordForm
	Tasks  []*Task
}

func (h *DashboardHandlers) HandleRecordsNew(w http.ResponseWriter, r *http.Request) {
	// time.Sleep(1 * time.Second)
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	// Set Task
	taskId := 0
	taskIdStr := r.URL.Query().Get("taskId")
	if taskIdStr != "" {
		taskId, _ = strconv.Atoi(taskIdStr)
		task := h.repo.TaskByID(taskId)
		if task == nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.UserID != user.ID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	// Set TimeStart
	now, _ := utils.NowWithTimezone(user.TimeZone)
	var timeStart string
	dateStr := r.URL.Query().Get("date")
	if dateStr != "" {
		timeStart = dateStr + "T" + now.Format("15:04")
	} else {
		timeStart = utils.FormatTimeForInput(&now)
	}

	form := recordForm{
		TaskID:    taskId,
		TimeStart: timeStart,
		TimeEnd:   "",
	}
	data := recordFormData{
		Form:   form,
		Errors: utils.FormErrors{},
		Tasks:  h.repo.Tasks(user.ID, ""),
	}
	h.renderRecordForm(w, data)
}

func (h *DashboardHandlers) HandleRecordsCreate(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	var form recordForm
	utils.ParseFormToStruct(r, &form)
	formErrors := utils.NewValidator(&form).Validate()
	if !formErrors.HasErrors() {
		task := h.repo.TaskByID(form.TaskID)
		if task == nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.UserID != user.ID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		h.validateIntersectingRecords(form, user, 0, formErrors)
		if !formErrors.HasErrors() {
			_, err := h.repo.CreateRecord(&Record{
				TaskID:    form.TaskID,
				TimeStart: *parseTimeFromInput(form.TimeStart),
				TimeEnd:   parseTimeFromInput(form.TimeEnd),
				Comment:   form.Comment,
			})

			if err == nil {
				w.Header().Set("HX-Trigger", "load-records, close-modal")
				w.Write([]byte("ok"))
				return
			} else {
				formErrors.Add("Common", "Error. Please try again later.")
			}
		}
	}

	data := recordFormData{
		Form:   form,
		Errors: formErrors,
		Tasks:  h.repo.Tasks(user.ID, ""),
	}
	h.renderRecordForm(w, data)
}

func (h *DashboardHandlers) HandleRecordsEdit(w http.ResponseWriter, r *http.Request) {
	user, record := h.getUserAndRecord(w, r)
	if user == nil || record == nil {
		return
	}

	form := recordForm{
		ID:        record.ID,
		TaskID:    record.TaskID,
		TimeStart: utils.FormatTimeForInput(&record.TimeStart),
		TimeEnd:   utils.FormatTimeForInput(record.TimeEnd),
		Comment:   record.Comment,
	}
	// List active tasks
	tasks := h.repo.Tasks(user.ID, "")
	if record.Task.IsCompleted {
		// Add current inactive task
		tasks = append(tasks, record.Task)
	}
	data := recordFormData{
		Form:   form,
		Errors: utils.FormErrors{},
		Tasks:  tasks,
	}
	h.renderRecordForm(w, data)
}

func (h *DashboardHandlers) HandleRecordsUpdate(w http.ResponseWriter, r *http.Request) {
	user, record := h.getUserAndRecord(w, r)
	if user == nil || record == nil {
		return
	}

	var form recordForm
	utils.ParseFormToStruct(r, &form)
	formErrors := utils.NewValidator(&form).Validate()
	if !formErrors.HasErrors() {
		h.validateIntersectingRecords(form, user, record.ID, formErrors)
		if !formErrors.HasErrors() {
			err := h.repo.UpdateRecord(&Record{
				ID:        record.ID,
				TaskID:    form.TaskID,
				TimeStart: *parseTimeFromInput(form.TimeStart),
				TimeEnd:   parseTimeFromInput(form.TimeEnd),
				Comment:   form.Comment,
			})
			if err == nil {
				w.Header().Set("HX-Trigger", "load-records, close-modal")
				w.Write([]byte(`ok`))
				return
			} else {
				formErrors.Add("Common", "Error. Please try again later.")
			}
		}
	}
	// List active tasks
	tasks := h.repo.Tasks(user.ID, "")
	if record.Task.IsCompleted {
		// Add current inactive task
		tasks = append(tasks, record.Task)
	}
	form.ID = record.ID
	data := recordFormData{
		Form:   form,
		Errors: formErrors,
		Tasks:  tasks,
	}
	h.renderRecordForm(w, data)
}

func (h *DashboardHandlers) HandleRecordsDelete(w http.ResponseWriter, r *http.Request) {
	user, record := h.getUserAndRecord(w, r)
	if user == nil || record == nil {
		return
	}
	h.repo.DeleteRecord(record.ID)
	w.Header().Set("HX-Trigger", "load-records")
	w.Write([]byte(`ok`))
}

func (h *DashboardHandlers) HandleRecordsList(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}
	week := r.URL.Query().Get("week")
	nowWithTimezone, _ := utils.NowWithTimezone(user.TimeZone)
	startInterval, endInterval := GetDateInterval(week, nowWithTimezone, user.IsWeekStartMonday)
	filterRecords := FilterRecords{
		UserID:        user.ID,
		StartInterval: startInterval,
		EndInterval:   endInterval,
	}
	// D("filterRecords r", "filterRecords", filterRecords)

	dailyRecords := h.repo.DailyRecords(filterRecords, nowWithTimezone)

	previousWeek := utils.FormatISOWeek(startInterval.AddDate(0, 0, -7), user.IsWeekStartMonday)
	nextWeek := utils.FormatISOWeek(endInterval.AddDate(0, 0, 7), user.IsWeekStartMonday)
	week = utils.FormatISOWeek(startInterval, user.IsWeekStartMonday)
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/record_list", "dashboard/record_list_navigation"}, "dashboard/record_list", utils.TplData{
		"DailyRecords":    dailyRecords,
		"User":            user,
		"Week":            week,
		"PreviousWeek":    previousWeek,
		"NextWeek":        nextWeek,
		"NowWithTimezone": nowWithTimezone,
	})
}

func (h *DashboardHandlers) renderRecordForm(w http.ResponseWriter, data recordFormData) {
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/record_form"}, "dashboard/record_form", utils.TplData{
		"Errors": data.Errors,
		"Form":   data.Form,
		"Tasks":  data.Tasks,
	})
}

func (h *DashboardHandlers) getUserAndRecord(w http.ResponseWriter, r *http.Request) (user *users.User, record *Record) {
	user = users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	recordIDStr := r.PathValue("id")
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil {
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	record = h.repo.RecordByIDWithTask(recordID)
	if record == nil {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	if record.Task.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	return
}

func (h *DashboardHandlers) validateIntersectingRecords(form recordForm, user *users.User, currentRecordId int, formErrors utils.FormErrors) {
	timeStart := parseTimeFromInput(form.TimeStart)
	timeEnd := parseTimeFromInput(form.TimeEnd)
	effectiveEnd := utils.EffectiveTime(timeEnd, user.TimeZone)

	if timeEnd != nil && (timeEnd.Before(*timeStart) || timeEnd.Equal(*timeStart)) {
		formErrors.Add("TimeEnd", "Time End must be greater than Time Start")
	}

	if timeEnd == nil {
		intersectingRecords := h.repo.RecordsWithTasks(FilterRecords{
			UserID:      user.ID,
			NotRecordID: currentRecordId,
			InProgress:  true,
		})
		if len(intersectingRecords) > 0 {
			message := "You are already doing task: " + recortToString(intersectingRecords[0], user)
			formErrors.Add("TimeEnd", message)
			return
		}
	}

	nowWithTimezone, _ := utils.NowWithTimezone(user.TimeZone)
	excludeInProgress := nowWithTimezone.Before(*timeStart)
	intersectingRecords := h.repo.RecordsWithTasks(FilterRecords{
		UserID:            user.ID,
		StartInterval:     *timeStart,
		EndInterval:       *effectiveEnd,
		NotRecordID:       currentRecordId,
		ExcludeInProgress: excludeInProgress,
	})
	if len(intersectingRecords) > 0 {
		message := "The selected time overlaps with other entries: "
		for _, record := range intersectingRecords {
			message += "<br> " + recortToString(record, user)
		}

		formErrors.Add("TimeEnd", message)
	}
}

func recortToString(record *Record, user *users.User) string {
	return fmt.Sprintf(
		"<a href=\"/dashboard?record=%d\" target=\"_blank\">%s %s %s</a>",
		record.ID,
		html.EscapeString(record.Task.Title),
		utils.FormatTimeRange(record.TimeStart, record.TimeEnd, user.TimeZone),
		html.EscapeString(record.Comment),
	)
}

// Time can be nil, so *time.Time
func parseTimeFromInput(input string) *time.Time {
	if input == "" {
		return nil
	}
	parsedTime, err := time.Parse("2006-01-02T15:04", input)
	if err != nil {
		return nil
	}
	return &parsedTime
}
