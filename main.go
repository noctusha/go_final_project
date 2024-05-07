package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/joho/godotenv"
	"github.com/noctusha/finalya/repeat"
)


/*func launchingServer() {
    err := http.ListenAndServe(os.Getenv("TODO_PORT"), http.FileServer(http.Dir("web")))
    if err != nil {
        log.Fatal("Error handling Listen and Serve")
    }
}
*/

func NextDateHandler(w http.ResponseWriter, r *http.Request) {

	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeatt := r.URL.Query().Get("repeat")

	// Проверка наличия параметров
	if now == "" || date == "" || repeatt == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		return
	}

	// Вызов функции NextDate из пакета repeat
	nextDate, err := repeat.NextDate(nowTime, date, repeatt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("nextDate"))
		return
	}

	// Отправка успешного ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

func connectingDB() {

	var dbFile string

    if os.Getenv("TODO_DBFILE") != "" {
        dbFile = os.Getenv("TODO_DBFILE")
    }  else {
    appPath, err := os.Executable() // путь к текущему файлу
    if err != nil {
        log.Fatal(err)
    }
    dbFile = filepath.Join(filepath.Dir(appPath), "scheduler.db") // приплюсовали к пути название ДБ
    }

    _, err := os.Stat(dbFile) // выдаст ошибку, если путь, собранный строкой выше, ни к чему не привел (дб нет)

    var install bool
    if err != nil {
        install = true // если ошибка была, значит инсталл будет тру и дб нужно создать (ниже)
    }

    db, err := sql.Open("sqlite", "scheduler.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if install {
        statement, err := db.Prepare(`CREATE TABLE scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date CHAR(8) NOT NULL DEFAULT "",
            title VARCHAR(256) NOT NULL DEFAULT "",
            comment VARCHAR(256) NOT NULL DEFAULT "",
            repeat VARCHAR(256) NOT NULL DEFAULT ""
            );
            CREATE INDEX tasks_date on tasks (date);
            `)

        if err != nil {
            log.Fatal(err)
        }

        statement.Exec()
    }

}

func main() {

	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    // Создаем новый маршрутизатор
    mux := http.NewServeMux()

    // Обработчик для API
    mux.HandleFunc("/api/nextdate", NextDateHandler)

    // Обработчик для статических файлов
    fs := http.FileServer(http.Dir("web"))
    mux.Handle("/", http.StripPrefix("/", fs))

    // Запускаем сервер с маршрутизатором
    fmt.Println("server is running")
    connectingDB()
    err = http.ListenAndServe(os.Getenv("TODO_PORT"), mux)
    if err != nil {
        log.Fatal("Error handling Listen and Serve")
    }

}