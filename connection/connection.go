package connection

import (
	"database/sql"
	"os"
	"time"

	"github.com/noctusha/finalya/models"
	_ "modernc.org/sqlite"
)

type Repository struct {
	DB *sql.DB
}

func ConnectingDB() (*Repository, error) {
	dbPath := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}

	return &Repository{DB: db}, nil
}

func (repo *Repository) Close() {
	repo.DB.Close()
}

func (repo *Repository) AddTask(task models.Task) (int64, error) {
	res, err := repo.DB.Exec("insert into scheduler (date, title, comment, repeat) values (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *Repository) GetTaskByID(id string) (models.Task, error) {
	var task models.Task
	row := repo.DB.QueryRow("select * from scheduler where id = ?", id)

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return models.Task{}, err
	}

	return task, nil
}

func (repo *Repository) UpdateTask(task models.Task) error {
	var tmp models.Task
	row := repo.DB.QueryRow("select * from scheduler where id = ?", task.ID)
	err := row.Scan(&tmp.ID, &tmp.Date, &tmp.Title, &tmp.Comment, &tmp.Repeat)

	if err != nil {
		return err
	}

	_, err = repo.DB.Exec("UPDATE scheduler set date = ?, title = ?, comment = ?, repeat = ? where id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) DeleteTask(id string) error {
	row := repo.DB.QueryRow("select * from scheduler where id = ?", id)
	var task models.Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return err
	}

	_, err = repo.DB.Exec("delete from scheduler where id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

func (repo *Repository) ListTasks(limit int, search string) ([]models.Task, error) {
	tasks := []models.Task{}
	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = repo.DB.Query("select id, date, title, comment, repeat from scheduler order by date limit ?", limit)
	} else {
		searchdate, err := time.Parse("02.01.2006", search)
		if err != nil {
			rows, err = repo.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?", "%"+search+"%", "%"+search+"%", limit)
		} else {
			correctsearchdate := searchdate.Format("20060102")
			rows, err = repo.DB.Query("select id, date, title, comment, repeat from scheduler where date = ?", correctsearchdate)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		task := models.Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}
