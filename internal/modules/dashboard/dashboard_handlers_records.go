package dashboard

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

type formRecord struct {
	taskID    int       `form:"task_id" validate:"required"`
	timeStart time.Time `form:"time_start" validate:"required,datetime=2006-01-02T15:04:05"`
	timeEnd   time.Time `form:"time_end" validate:"datetime=2006-01-02T15:04:05"`
	comment   string    `form:"comment" validate:"max=10000"`
}

func (h *DashboardHandlers) renderRecordForm(w http.ResponseWriter, form formRecord, formErrors utils.FormErrors, url string) {
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/form_record"}, "dashboard/form_record", utils.TplData{
		"Errors": formErrors,
		"Form":   form,
		"URL":    url,
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

func (h *DashboardHandlers) HandleRecordsNew(w http.ResponseWriter, r *http.Request) {
	// time.Sleep(1 * time.Second)
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}
	h.renderRecordForm(w, formRecord{timeStart: time.Now()}, utils.FormErrors{}, "/tasks")
}

func (h *DashboardHandlers) HandleRecordsCreate(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	var form formRecord
	utils.ParseFormToStruct(r, &form)
	formErrors := utils.NewValidator(&form).Validate()
	if !formErrors.HasErrors() {
		task := h.repo.TaskByID(form.taskID)
		if task == nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if task.UserID != user.ID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		h.repo.CreateRecord(&Record{
			TaskID:    form.taskID,
			TimeStart: form.timeStart,
			TimeEnd:   form.timeEnd,
			Comment:   form.comment,
		})

		w.Header().Set("HX-Trigger", "load-records, close-modal")
		w.Write([]byte("ok"))
		return
	}

	h.renderRecordForm(w, form, formErrors, "/records")
}

func (h *DashboardHandlers) HandleRecordsEdit(w http.ResponseWriter, r *http.Request) {
	user, record := h.getUserAndRecord(w, r)
	if user == nil || record == nil {
		return
	}

	form := formRecord{
		taskID:    record.TaskID,
		timeStart: record.TimeStart,
		timeEnd:   record.TimeEnd,
		comment:   record.Comment,
	}
	h.renderRecordForm(w, form, utils.FormErrors{}, fmt.Sprintf("/records/%d", record.ID))
}

func (h *DashboardHandlers) HandleRecordsUpdate(w http.ResponseWriter, r *http.Request) {
	user, record := h.getUserAndRecord(w, r)
	if user == nil || record == nil {
		return
	}

	var form formRecord
	utils.ParseFormToStruct(r, &form)
	formErrors := utils.NewValidator(&form).Validate()
	if !formErrors.HasErrors() {
		err := h.repo.UpdateRecord(&Record{
			ID:        record.ID,
			TaskID:    form.taskID,
			TimeStart: form.timeStart,
			TimeEnd:   form.timeEnd,
			Comment:   form.comment,
		})
		if err == nil {
			w.Header().Set("HX-Trigger", "load-records, close-modal")
			w.Write([]byte(`ok`))
			return
		}
	}

	h.renderRecordForm(w, form, formErrors, fmt.Sprintf("/records/%d", record.ID))
}

func (h *DashboardHandlers) HandleRecordsDelete(w http.ResponseWriter, r *http.Request) {
	user, record := h.getUserAndRecord(w, r)
	if user == nil || record == nil {
		return
	}
	h.repo.DeleteTask(record.ID)
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
