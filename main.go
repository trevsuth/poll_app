package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var ctx = context.Background()
var redisClient *redis.Client

type Poll struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/create", createPollHandler).Methods("GET", "POST")
	r.HandleFunc("/voting", votingHandler)
	r.HandleFunc("/vote", voteHandler).Methods("POST")
	r.HandleFunc("/admin", adminHandler)
	r.HandleFunc("/results", resultsHandler).Methods("POST")
	r.HandleFunc("/reset", resetHandler).Methods("POST")
	http.Handle("/", r)

	fmt.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func getPollHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pollId := vars["pollId"]
	pollData, err := redisClient.Get(ctx, pollId).Result()
	if err != nil {
		http.Error(w, "Poll not found", 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(pollData))
}

// Submit Vote
func submitVoteHandler(w http.ResponseWriter, r *http.Request) {
	var vote struct {
		PollID string `json:"pollId"`
		Answer string `json:"answer"`
	}
	json.NewDecoder(r.Body).Decode(&vote)
	// Increment vote count for the given answer in Redis
	redisClient.HIncrBy(ctx, "poll:"+vote.PollID, vote.Answer, 1)
	fmt.Fprintf(w, "Vote recorded successfully for %s", vote.Answer)
}

func createPollHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var poll struct {
			PollID   string   `json:"pollId"`
			Question string   `json:"question"`
			Answers  []string `json:"answers"`
		}
		err := json.NewDecoder(r.Body).Decode(&poll)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		pollData, _ := json.Marshal(poll)
		// Use PollID as the key in Redis
		redisClient.Set(ctx, poll.PollID, pollData, 0)
		fmt.Fprintf(w, "Poll created successfully with ID: %s", poll.PollID)
	} else {
		http.ServeFile(w, r, "templates/create.html")
	}
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
