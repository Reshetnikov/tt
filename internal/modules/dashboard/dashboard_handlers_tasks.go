package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time-tracker/internal/modules/users"
	"time-tracker/internal/utils"
)

type formTask struct {
	Title       string `form:"title" validate:"required,min=1,max=255"`
	Description string `form:"description" validate:"max=10000"`
	Color       string `form:"color" validate:"required,hexcolor"`
	IsCompleted bool   `form:"is_completed" label:"Completed"`
}

func (h *DashboardHandlers) HandleTasksNew(w http.ResponseWriter, r *http.Request) {
	// time.Sleep(1 * time.Second)
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}
	h.renderTaskForm(w, formTask{Color: "#FFFFFF"}, utils.FormErrors{}, "/tasks")
}

func (h *DashboardHandlers) HandleTasksCreate(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	var form formTask
	utils.ParseFormToStruct(r, &form)
	formErrors := utils.NewValidator(&form).Validate()
	if !formErrors.HasErrors() {
		h.repo.CreateTask(&Task{
			UserID:      user.ID,
			Title:       form.Title,
			Description: form.Description,
			Color:       form.Color,
			IsCompleted: form.IsCompleted,
		})

		w.Header().Set("HX-Trigger", "load-tasks, close-modal")
		w.Write([]byte("ok"))
		return
	}

	h.renderTaskForm(w, form, formErrors, "/tasks")
}

func (h *DashboardHandlers) HandleTasksEdit(w http.ResponseWriter, r *http.Request) {
	user, task := h.getUserAndTask(w, r)
	if user == nil || task == nil {
		return
	}

	form := formTask{
		Title:       task.Title,
		Description: task.Description,
		Color:       task.Color,
		IsCompleted: task.IsCompleted,
	}
	h.renderTaskForm(w, form, utils.FormErrors{}, fmt.Sprintf("/tasks/%d", task.ID))
}

func (h *DashboardHandlers) HandleTasksUpdate(w http.ResponseWriter, r *http.Request) {
	user, task := h.getUserAndTask(w, r)
	if user == nil || task == nil {
		return
	}

	var form formTask
	utils.ParseFormToStruct(r, &form)
	formErrors := utils.NewValidator(&form).Validate()
	if !formErrors.HasErrors() {
		if form.IsCompleted != task.IsCompleted {
			task.SortOrder = h.repo.GetMaxSortOrder(user.ID, form.IsCompleted) + 1
		}
		err := h.repo.UpdateTask(&Task{
			ID:          task.ID,
			Title:       form.Title,
			Description: form.Description,
			Color:       form.Color,
			IsCompleted: form.IsCompleted,
			SortOrder:   task.SortOrder,
		})
		if err == nil {
			w.Header().Set("HX-Trigger", "load-tasks, load-records, close-modal")
			w.Write([]byte(`ok`))
			return
		}
	}

	h.renderTaskForm(w, form, formErrors, fmt.Sprintf("/tasks/%d", task.ID))
}

func (h *DashboardHandlers) HandleTasksDelete(w http.ResponseWriter, r *http.Request) {
	user, task := h.getUserAndTask(w, r)
	if user == nil || task == nil {
		return
	}
	h.repo.DeleteTask(task.ID)
	w.Header().Set("HX-Trigger", "load-tasks, load-records")
	w.Write([]byte(`ok`))
}

func (h *DashboardHandlers) HandleTaskList(w http.ResponseWriter, r *http.Request) {
	user := users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}
	taskCompleted := r.URL.Query().Get("taskCompleted")
	tasks := h.repo.Tasks(user.ID, taskCompleted)
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/task_list"}, "dashboard/task_list", utils.TplData{
		"Tasks":         tasks,
		"TaskCompleted": taskCompleted,
	})
}

func (h *DashboardHandlers) HandleUpdateSortOrder(w http.ResponseWriter, r *http.Request) {
	// bytedata, err := io.ReadAll(r.Body)
	// reqBodyString := string(bytedata)
	// D("Decode", "reqBodyString", reqBodyString)

	user := users.GetUserFromRequest(r)
	if user == nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	var order []struct {
		ID        int `json:"id"`
		SortOrder int `json:"sortOrder"`
	}

	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	for _, task := range order {
		err := h.repo.UpdateTaskSortOrder(task.ID, user.ID, task.SortOrder)
		if err != nil {
			http.Error(w, "Error updating task order", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	// The request is sent via fetch, not html, so tiger must be called in js: htmx.trigger(document.body, "load-tasks");
	// w.Header().Set("HX-Trigger", "load-tasks")
	w.Write([]byte(`{"status": "success"}`))
}

func (h *DashboardHandlers) renderTaskForm(w http.ResponseWriter, form formTask, formErrors utils.FormErrors, url string) {
	utils.RenderTemplateWithoutLayout(w, []string{"dashboard/task_form"}, "dashboard/task_form", utils.TplData{
		"Errors": formErrors,
		"Form":   form,
		"URL":    url,
	})
}

func (h *DashboardHandlers) getUserAndTask(w http.ResponseWriter, r *http.Request) (user *users.User, task *Task) {
	user = users.GetUserFromRequest(r)
	if user == nil {
		utils.RenderBlockNeedLogin(w)
		return
	}

	taskIDStr := r.PathValue("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task = h.repo.TaskByID(taskID)
	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if task.UserID != user.ID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	return
}
