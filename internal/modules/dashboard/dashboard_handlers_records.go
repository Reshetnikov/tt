package dashboard

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

type recordForm struct {
	TaskID    int    `form:"task_id" validate:"required"`
	TimeStart string `form:"time_start" validate:"required,datetime=2006-01-02T15:04" label:"Time Start"`
	TimeEnd   string `form:"time_end" validate:"omitempty,datetime=2006-01-02T15:04" label:"Time End"`
	Comment   string `form:"comment" validate:"max=10000"`
}

type recordFormData struct {
	Errors utils.FormErrors
	Form   recordForm
	URL    string
	Tasks  []*Task
}

func (h *DashboardHandlers) HandleRecordsNew(w http.ResponseWriter, r *http.Request) {
	// time.Sleep(1 * time.Second)
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	taskId := 0
	taskIdStr := r.URL.Query().Get("taskId")
	if taskIdStr != "" {
		taskId, _ = strconv.Atoi(taskIdStr)
	}

	now, _ := utils.NowWithTimezone(user.TimeZone)
	form := recordForm{
		TaskID:    taskId,
		TimeStart: formatTimeForInput(&now),
		TimeEnd:   "",
	}
	data := recordFormData{
		Form:   form,
		Errors: utils.FormErrors{},
		URL:    "/records",
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
		h.repo.CreateRecord(&Record{
			TaskID:    form.TaskID,
			TimeStart: *parseTimeFromInput(form.TimeStart),
			TimeEnd:   parseTimeFromInput(form.TimeEnd),
			Comment:   form.Comment,
		})

		w.Header().Set("HX-Trigger", "load-records, close-modal")
		w.Write([]byte("ok"))
		return
	}

	data := recordFormData{
		Form:   form,
		Errors: formErrors,
		URL:    "/records",
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
		TaskID:    record.TaskID,
		TimeStart: formatTimeForInput(&record.TimeStart),
		TimeEnd:   formatTimeForInput(record.TimeEnd),
		Comment:   record.Comment,
	}
	tasks := h.repo.Tasks(user.ID, "") // Active tasks
	if record.Task.IsCompleted {
		tasks = append(tasks, record.Task) // Add current inactive task
	}
	data := recordFormData{
		Form:   form,
		Errors: utils.FormErrors{},
		URL:    fmt.Sprintf("/records/%d", record.ID),
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
		}
	}
	tasks := h.repo.Tasks(user.ID, "") // Active tasks
	if record.Task.IsCompleted {
		tasks = append(tasks, record.Task) // Add current inactive task
	}
	data := recordFormData{
		Form:   form,
		Errors: formErrors,
		URL:    fmt.Sprintf("/records/%d", record.ID),
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
	records := h.repo.RecordsWithTasks(FilterRecords{
		UserID: user.ID,
		// RecordID: 0,
		// Start:    time.Now().Add(-7 * 24 * time.Hour),
		// End:      time.Now(),
	})
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/record_list"}, "dashboard/record_list", utils.TplData{
		"Records": records,
	})
}

func (h *DashboardHandlers) renderRecordForm(w http.ResponseWriter, data recordFormData) {
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/record_form"}, "dashboard/record_form", utils.TplData{
		"Errors": data.Errors,
		"Form":   data.Form,
		"URL":    data.URL,
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

func formatTimeForInput(t *time.Time) string {
	if t == nil {
		return "" // Пустое значение для nil
	}
	return t.Format("2006-01-02T15:04")
}
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
