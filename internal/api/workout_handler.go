package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ruhan/internal/middleware"
	"github.com/ruhan/internal/store"
	"github.com/ruhan/internal/utils"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
	logger       *log.Logger
}

func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *log.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore,
		logger,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutById(res http.ResponseWriter, req *http.Request) {
	workoutId, err := utils.ReadIdParam(req)

	if err != nil {
		wh.logger.Printf("ERROR: ReadIDParam: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "Invlaid workout id"})
	}

	workout, err := wh.workoutStore.GetWorkoutByID(workoutId)

	if err != nil {
		wh.logger.Printf("ERROR: GetWorkoutByID: %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	utils.WriteJSON(res, http.StatusOK, utils.Envelope{"workout": workout})
}

func (wh *WorkoutHandler) HandleCreateOut(res http.ResponseWriter, req *http.Request) {
	var workout store.Workout

	err := json.NewDecoder(req.Body).Decode(&workout)

	if err != nil {
		wh.logger.Printf("ERROR: Decoder: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "Failed to create workout"})
		return
	}

	currentUser := middleware.GetUser(req)
	if currentUser == nil || currentUser == store.AnonymousUser {
		wh.logger.Printf("ERROR: GetUser: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "You must be logged in"})
		return
	}

	workout.UserID = currentUser.ID

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)

	if err != nil {
		wh.logger.Printf("ERROR: CreateWorkout: %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create workout"})
		return
	}

	utils.WriteJSON(res, http.StatusCreated, utils.Envelope{"workout": createdWorkout})
}

func (wh *WorkoutHandler) HandleUpdateWorkoutById(res http.ResponseWriter, req *http.Request) {
	workoutId, err := utils.ReadIdParam(req)

	if err != nil {
		wh.logger.Printf("ERROR: ReadIDParam: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "Invlaid workout id"})
		return
	}

	existingWorkout, err := wh.workoutStore.GetWorkoutByID(workoutId)
	if err != nil {
		wh.logger.Printf("ERROR: GetWorkoutByID: %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}

	if existingWorkout == nil {
		utils.WriteJSON(res, http.StatusNotFound, utils.Envelope{"error": "Workout not found"})
		return
	}

	var updateWorkoutReq struct {
		Title           *string              `json:"title"`
		Description     *string              `json:"description"`
		DurationMinutes *int                 `json:"duration_minutes"`
		CaloriesBurned  *int                 `json:"calories_minutes"`
		Entries         []store.WorkoutEntry `json:"entries"`
	}

	err = json.NewDecoder(req.Body).Decode(&updateWorkoutReq)

	if err != nil {
		wh.logger.Printf("ERROR: Decoding: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	if updateWorkoutReq.Title != nil {
		existingWorkout.Title = *updateWorkoutReq.Title
	}

	if updateWorkoutReq.Description != nil {
		existingWorkout.Description = *updateWorkoutReq.Description
	}

	if updateWorkoutReq.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updateWorkoutReq.DurationMinutes
	}

	if updateWorkoutReq.Entries != nil {
		existingWorkout.Entries = updateWorkoutReq.Entries
	}

	currentUser := middleware.GetUser(req)
	if currentUser == nil || currentUser == store.AnonymousUser {
		wh.logger.Printf("ERROR: GetUser: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "You must be logged in"})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			wh.logger.Printf("ERROR: GetWorkoutOwner: %v", err)
			utils.WriteJSON(res, http.StatusNotFound, utils.Envelope{"error": "workout doesn't exists"})
			return
		}
		wh.logger.Printf("ERROR: GetWorkoutOwner Internal Server Error: %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	if workoutOwner != currentUser.ID {
		utils.WriteJSON(res, http.StatusForbidden, utils.Envelope{"error": "you are not authorized to update it"})
		return
	}

	err = wh.workoutStore.UpdateWorkout(existingWorkout)

	if err != nil {
		wh.logger.Printf("ERROR: UpdateWorkout: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	utils.WriteJSON(res, http.StatusOK, utils.Envelope{"workout": existingWorkout})
}

func (wh *WorkoutHandler) HandleDeleteWorkoutById(res http.ResponseWriter, req *http.Request) {
	workoutId, err := utils.ReadIdParam(req)

	if err != nil {
		wh.logger.Printf("ERROR: ReadIDParam: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "Invlaid workout id"})
	}

	currentUser := middleware.GetUser(req)
	if currentUser == nil || currentUser == store.AnonymousUser {
		wh.logger.Printf("ERROR: GetUser: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "You must be logged in"})
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			wh.logger.Printf("ERROR: GetWorkoutOwner: %v", err)
			utils.WriteJSON(res, http.StatusNotFound, utils.Envelope{"error": "workout doesn't exists"})
			return
		}
		wh.logger.Printf("ERROR: GetWorkoutOwner Internal Server Error: %v", err)
		utils.WriteJSON(res, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	if workoutOwner != currentUser.ID {
		utils.WriteJSON(res, http.StatusForbidden, utils.Envelope{"error": "you are not authorized to delete it"})
		return
	}

	err = wh.workoutStore.DeleteWorkout(workoutId)

	if err != nil {
		wh.logger.Printf("ERROR: DeleteWorkout: %v", err)
		utils.WriteJSON(res, http.StatusBadRequest, utils.Envelope{"error": "Failed to delete workout"})
		return
	}

	utils.WriteJSON(res, http.StatusOK, utils.Envelope{"success": true})
}
