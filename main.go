package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/ruhan/internal/app"
	"github.com/ruhan/internal/routes"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8001, "this backend server port")
	flag.Parse()

	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}

	defer app.DB.Close()

	routesHandler := routes.SetupRoutes(app)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      routesHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.Logger.Printf("App is running on Port: %d\n", port)
	err = server.ListenAndServe()

	if err != nil {
		app.Logger.Fatal(err)
	}
}
