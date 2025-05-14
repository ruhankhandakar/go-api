package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ruhan/internal/app"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/health", app.HealthCheck)
	r.Get("/workouts/{id}", app.WorkoutHandler.HandleGetWorkoutById)

	r.Post("/workouts", app.WorkoutHandler.HandleCreateOut)
	r.Put("/workouts/{id}", app.WorkoutHandler.HandleUpdateWorkoutById)
	r.Delete("/workouts/{id}", app.WorkoutHandler.HandleDeleteWorkoutById)

	r.Post("/users", app.UserHandler.HandleRegisterUser)

	r.Post("/tokens/authentication", app.TokenHandler.HandleCreateToken)

	return r
}
