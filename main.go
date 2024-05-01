package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var ctx = context.Background()
var redisClient *redis.Client

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // Default DB
	})

	// Set up routes
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/voting", votingHandler)
	r.HandleFunc("/admin", adminHandler)
	r.HandleFunc("/vote", voteHandler).Methods("POST")
	r.HandleFunc("/results", resultsHandler).Methods("GET")
	r.HandleFunc("/reset", resetHandler).Methods("POST")
	http.Handle("/", r)

	// Start server
	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func votingHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/voting.html")
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	// Reset vote counts in Redis
	redisClient.Set(ctx, "yes", 0, 0)
	redisClient.Set(ctx, "no", 0, 0)
	fmt.Fprint(w, "Votes reset successfully")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/admin.html")
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	vote := r.FormValue("vote")
	redisClient.Incr(ctx, vote)
	fmt.Fprint(w, "Vote registered")
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	yesVotes, err := redisClient.Get(ctx, "yes").Result()
	if err != nil {
		yesVotes = "0"
	}
	noVotes, err := redisClient.Get(ctx, "no").Result()
	if err != nil {
		noVotes = "0"
	}
	fmt.Fprintf(w, "Results - Yes: %s, No: %s", yesVotes, noVotes)
}
