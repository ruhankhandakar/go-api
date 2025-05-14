package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ruhan/internal/api"
	"github.com/ruhan/internal/store"
	"github.com/ruhan/migrations"
)

type Application struct {
	Logger         *log.Logger
	WorkoutHandler *api.WorkoutHandler
	UserHandler    *api.UserHandler
	DB             *sql.DB
}

func NewApplication() (*Application, error) {
	logger := log.New(os.Stdout, "GO: ", log.Ldate|log.Ltime)

	// stores
	pgDb, err := store.Open()
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(pgDb, migrations.FS, ".")

	if err != nil {
		panic(err)
	}

	// our stores
	workoutStore := store.NewPostgresWorkoutStore(pgDb)
	userStore := store.NewPostgreUserStore(pgDb)

	// Handlers
	workoutHandler := api.NewWorkoutHandler(workoutStore, logger)
	userHandler := api.NewUserHandler(userStore, logger)

	app := &Application{
		Logger:         logger,
		WorkoutHandler: workoutHandler,
		UserHandler:    userHandler,
		DB:             pgDb,
	}

	return app, nil
}

func (app *Application) HealthCheck(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "Status is available")
}
