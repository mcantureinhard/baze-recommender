package main

import (
	"log"
	"net/http"
	"pill-recommender/service/src/recommendation"

	"github.com/gorilla/handlers"
)

func main() {
	// Get the "PORT" env variable
	//port := os.Getenv("PORT")
	port := "8082"

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := recommendation.NewRouter() // create routes

	// These two lines are important if you're designing a front-end to utilise this API methods
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT"})

	// Launch server with CORS validations
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(allowedOrigins, allowedMethods)(router)))
}
