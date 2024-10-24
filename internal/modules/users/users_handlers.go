package users

import (
	"fmt"
	"net/http"
	"time-tracker/internal/utils"

	"github.com/go-playground/validator/v10"
)

type UsersHandler struct {
	usersService *UsersService 
	validate *validator.Validate
}

func NewUsersHandlers(usersService *UsersService) *UsersHandler {
	return &UsersHandler{
		usersService: usersService, 
		validate: validator.New(),
	}
}

// Обработчик для регистрации
type signupForm struct {
	Name                 string `form:"name" validate:"required,min=2,max=40"`
	Email                string `form:"email" validate:"required,email"`
	Password             string `form:"password" validate:"required,min=8"`
	PasswordConfirmation string `form:"password_confirmation" validate:"required,eqfield=Password"`
}

type FormErrors map[string][]string
func (fe *FormErrors) Add(field, message string) {
	(*fe)[field] = append((*fe)[field], message)
}
func (fe FormErrors) HasErrors(field string) bool {
	return len(fe[field]) > 0
}

func parseValidationError(tag string, err validator.FieldError) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", err.Field(), err.Param())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}

func (h *UsersHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("-----%+v\n", r)
	if r.Method == http.MethodPost {

		form, err := utils.ParseFormToStruct[signupForm](r)
		if err != nil {
			http.Error(w, "Unable to parse form data", http.StatusBadRequest)
			return
		}

		errors := FormErrors{}
		if err := h.validate.Struct(form); err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				field := err.Field() // Name
				tag := err.Tag() // min
				errorMessage := parseValidationError(tag, err)
				errors.Add(field, errorMessage)
			}

			// Передаем ошибки в шаблон
			utils.RenderTemplate(w, "signup", map[string]interface{}{
				"Title":  "Sign Up",
				"Errors": errors,
				"Form":   form, // Передаем данные формы, чтобы они сохранились в полях
			})
			return
		} else {
			utils.RenderTemplate(w, "signup", map[string]interface{}{
				"Title":  "Sign Up",
				"Errors": errors,
				"Form":   form, // Передаем данные формы, чтобы они сохранились в полях
			})
		}
	}
	utils.RenderTemplate(w, "signup", map[string]interface{}{
		"Title":  "Sign Up",
	})
}

// Обработчик для входа
func (h *UsersHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Логика входа пользователя с использованием userService
}
