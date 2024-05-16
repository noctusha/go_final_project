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

    // Создаем новый маршрутизатор
    mux := http.NewServeMux()

    // Обработчик для API
    mux.HandleFunc("/api/nextdate", handlers.NextDateHandler)
    mux.HandleFunc("/api/task", handlers.TaskHandler)
	mux.HandleFunc("/api/tasks", handlers.ListTasksHandler)
	mux.HandleFunc("/api/task/done", handlers.DoneTaskHandler)

    // Обработчик для статических файлов
    fs := http.FileServer(http.Dir("web"))
    mux.Handle("/", http.StripPrefix("/", fs))

    fmt.Println("server is running")

    db := connection.ConnectingDB()
	defer db.Close()

    err = http.ListenAndServe(os.Getenv("TODO_PORT"), mux)
    if err != nil {
        log.Fatal("Error handling Listen and Serve")
	}
}