package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/joho/godotenv"
)

func launchingServer() {
    err := http.ListenAndServe(os.Getenv("TODO_PORT"), http.FileServer(http.Dir("web")))
    if err != nil {
        log.Fatal("Error handling Listen and Serve")
    }
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

    fmt.Println("server is running")
    connectingDB()
    launchingServer()
}
