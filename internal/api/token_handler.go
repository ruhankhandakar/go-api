package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ruhan/internal/store"
	"github.com/ruhan/internal/tokens"
	"github.com/ruhan/internal/utils"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

type createTokenRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore,
		userStore,
		logger,
	}
}

func (h *TokenHandler) HandleCreateToken(res http.ResponseWriter, req *http.Request) {
	var body createTokenRequest

	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		h.logger.Printf("ERROR: create token request %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "invalid payload"})
		return
	}

	user, err := h.userStore.GetUserByUserName(body.UserName)
	if err != nil {
		h.logger.Printf("ERROR: GetUserByUserName %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	passwordDoMatch, err := user.PasswordHash.Matches(body.Password)
	if err != nil {
		h.logger.Printf("ERROR: Password hash match %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	if !passwordDoMatch {
		utils.WriteJSON(res, http.StatusUnauthorized, utils.Envelope{"error": "invalid credentials"})
		return
	}

	token, err := h.tokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		h.logger.Printf("ERROR: CreateNewToken %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	utils.WriteJSON(res, http.StatusCreated, utils.Envelope{"auth_token": token})
}
