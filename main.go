package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"github.com/joho/godotenv"
	"github.com/noctusha/finalya/connection"
	"github.com/noctusha/finalya/handlers"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	repo, err := connection.ConnectingDB()
    if err != nil {
        log.Fatal("Failed to connect to database")
    }
    defer repo.Close()

    handler := &handlers.Handler{Repo: repo}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	mux.HandleFunc("/api/task", handler.TaskHandler)
	mux.HandleFunc("/api/tasks", handler.ListTasksHandler)
	mux.HandleFunc("/api/task/done", handler.DoneTaskHandler)

	fs := http.FileServer(http.Dir("web"))
	mux.Handle("/", http.StripPrefix("/", fs))

	fmt.Println("server is running")

	err = http.ListenAndServe(os.Getenv("TODO_PORT"), mux)
	if err != nil {
		log.Fatal("Error handling Listen and Serve")
	}
}