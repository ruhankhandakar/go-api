package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/ruhan/internal/store"
	"github.com/ruhan/internal/utils"
)

type UserHandler struct {
	usreStore store.UserStore
	logger    *log.Logger
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore,
		logger,
	}
}

type registerUserRequest struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

func (h *UserHandler) validateRegisterRequest(req *registerUserRequest) error {
	if req.UserName == "" {
		return errors.New("username is required")
	}

	if len(req.UserName) > 50 {
		return errors.New("username cannot be greater than 50 characters")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email provided")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

func (h *UserHandler) HandleRegisterUser(res http.ResponseWriter, req *http.Request) {
	var reg registerUserRequest

	err := json.NewDecoder(req.Body).Decode(&reg)
	if err != nil {
		h.logger.Printf("ERROR: decoding register request %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "invalid payload"})
		return
	}

	err = h.validateRegisterRequest(&reg)
	if err != nil {
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	user := &store.User{
		UserName: reg.UserName,
		Email:    reg.Email,
	}

	if reg.Bio != "" {
		user.Bio = reg.Bio
	}

	err = user.PasswordHash.Set(reg.Password)
	if err != nil {
		h.logger.Printf("ERROR: hashing password %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "invalid authentication"})
		return
	}

	err = h.usreStore.CreateUser(user)
	if err != nil {
		h.logger.Printf("ERROR: registering user %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "issue with registering user"})
		return
	}

	utils.WriteJSON(res, http.StatusOK, utils.Envelope{"user": user})
}
