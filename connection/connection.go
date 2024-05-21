package connection

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func ConnectingDB() *sql.DB {
	dbPath := os.Getenv("TODO_DBFILE") // Получаем путь к файлу базы данных из переменной окружения

	// Попытка открыть существующий файл базы данных или создать новый, если он не существует
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	statement, err := db.Prepare(`CREATE TABLE if not exists scheduler (
            "id" INTEGER PRIMARY KEY AUTOINCREMENT,
            "date" CHAR(8) NOT NULL DEFAULT "",
            "title" VARCHAR(256) NOT NULL DEFAULT "",
            "comment" VARCHAR(256) NOT NULL DEFAULT "",
            "repeat" VARCHAR(256) NOT NULL DEFAULT ""
        );
        CREATE INDEX if not exists scheduler_date on scheduler (date);`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatalf("Failed to execute table creation statement: %v", err)
	}

	return db
}
