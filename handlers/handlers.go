package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	_ "modernc.org/sqlite"

	"github.com/noctusha/finalya/connection"
	"github.com/noctusha/finalya/repeatRule"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type JSON struct {
	ID    int64   `json:"id,omitempty"`
	Err   string  `json:"error,omitempty"`
	Tasks *[]Task `json:"tasks,omitempty"`
}

func respondJSON(w http.ResponseWriter, payload interface{}, statusCode int) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func respondJSONError(w http.ResponseWriter, message string, statusCode int) {
	respondJSON(w, JSON{Err: message}, statusCode)
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// смотрим наличие пароля
		passENV := os.Getenv("TODO_PASSWORD")
		if len(passENV) > 0 {
			var jwtB string // JWT-токен из куки
			// получаем куку
			cookie, err := r.Cookie("token")
			if err == nil {
				jwtB = cookie.Value
			}
			var valid bool
			// здесь код для валидации и проверки JWT-токена
			// ...

			var buf bytes.Buffer
			var pass string

			_, err = buf.ReadFrom(r.Body)
			if err != nil {
				respondJSONError(w, "Failed to read password", http.StatusBadRequest)
				return
			}

			err = json.Unmarshal(buf.Bytes(), &pass)
			if err != nil {
				respondJSONError(w, "Invalid JSON format", http.StatusBadRequest)
				return
			}

			jwtToken := jwt.New(jwt.SigningMethodHS256)

			signedToken, err := jwtToken.SignedString(pass)
			if err != nil {
				respondJSONError(w, "Failed to sign jwt", http.StatusBadRequest)
				return
			}

			if jwtB != signedToken {
				valid = false
			}

			if !valid {
				respondJSONError(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := r.URL.Query().Get("now")
	date := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if now == "" || date == "" || repeat == "" {
		respondJSONError(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		respondJSONError(w, "Wrong date format", http.StatusBadRequest)
		return
	}

	nextDate, err := repeatRule.NextDate(nowTime, date, repeat)
	if err != nil {
		respondJSONError(w, "Invalid repetition rate", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

func NewTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var task Task

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		respondJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		respondJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		respondJSONError(w, "Missing title", http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	dateTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		respondJSONError(w, "Wrong date format", http.StatusBadRequest)
		return
	}

	// проверяем разницу дат без учета часов/минут/секунд
	if dateTime.Before(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())) {
		if task.Repeat == "" {
			task.Date = time.Now().Format("20060102")
		} else {
			task.Date, err = repeatRule.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				respondJSONError(w, "Invalid repetition rate", http.StatusBadRequest)
				return
			}
		}
	}

	db := connection.ConnectingDB()
	defer db.Close()

	res, err := db.Exec("insert into scheduler (date, title, comment, repeat) values (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))

	if err != nil {
		respondJSONError(w, "Failed to insert task into database", http.StatusBadRequest)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		respondJSONError(w, "Failed to retrieve last insert ID", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{ID: id}, http.StatusOK)
}

func ChangeTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	id := r.URL.Query().Get("id")

	db := connection.ConnectingDB()
	defer db.Close()

	row := db.QueryRow("select * from scheduler where id = ?", id)

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		respondJSONError(w, "Failed to scan selected result from database", http.StatusBadRequest)
		return
	}

	respondJSON(w, task, http.StatusOK)
}

func PushChangedTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var task Task

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		respondJSONError(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &task)
	if err != nil {
		respondJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		respondJSONError(w, "Missing title", http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}

	dateTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		respondJSONError(w, "Wrong date format", http.StatusBadRequest)
		return
	}

	// проверяем разницу дат без учета часов/минут/секунд
	if dateTime.Before(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())) {
		if task.Repeat == "" {
			task.Date = time.Now().Format("20060102")
		} else {
			task.Date, err = repeatRule.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				respondJSONError(w, "Invalid repetition rate", http.StatusBadRequest)
				return
			}
		}
	}

	db := connection.ConnectingDB()
	defer db.Close()

	row := db.QueryRow("select * from scheduler where id = ?", task.ID)
	var tmp Task
	err = row.Scan(&tmp.ID, &tmp.Date, &tmp.Title, &tmp.Comment, &tmp.Repeat)

	if err != nil {
		respondJSONError(w, "Task not found", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE scheduler set date = ?, title = ?, comment = ?, repeat = ? where id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		respondJSONError(w, "Failed to update new data", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{}, http.StatusOK)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	db := connection.ConnectingDB()
	defer db.Close()

	row := db.QueryRow("select * from scheduler where id = ?", id)
	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		respondJSONError(w, "Failed to scan selected result from database", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("delete from scheduler where id = ?", id)
	if err != nil {
		respondJSONError(w, "Failed to delete selected task", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{}, http.StatusOK)
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		NewTask(w, r)

	case http.MethodGet:
		ChangeTask(w, r)

	case http.MethodPut:
		PushChangedTask(w, r)

	case http.MethodDelete:
		DeleteTask(w, r)

	default:
		return
	}
}

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks := []Task{}
	db := connection.ConnectingDB()
	defer db.Close()

	search := r.URL.Query().Get("search")

	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = db.Query("select id, date, title, comment, repeat from scheduler order by date limit :limit", sql.Named("limit", 20))
	} else {
		searchdate, err := time.Parse("02.01.2006", search)
		if err != nil {
			rows, err = db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?", "%"+search+"%", "%"+search+"%", 20)
		} else {
			correctsearchdate := searchdate.Format("20060102")
			rows, err = db.Query("select id, date, title, comment, repeat from scheduler where date = ?", correctsearchdate)
		}
	}

	if err != nil {
		respondJSONError(w, "Failed to select task from database", http.StatusBadRequest)
		return
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			respondJSONError(w, "Failed to scan selected result from database", http.StatusBadRequest)
			return
		}
		tasks = append(tasks, task)
	}

	err = rows.Err()
	if err != nil {
		respondJSONError(w, "Failed during rows iteration", http.StatusInternalServerError)
		return
	}

	respondJSON(w, JSON{Tasks: &tasks}, http.StatusOK)
}

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		id := r.URL.Query().Get("id")

		db := connection.ConnectingDB()
		defer db.Close()

		row := db.QueryRow("select * from scheduler where id = ?", id)
		var task Task
		err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			respondJSONError(w, "Failed to scan selected result from database", http.StatusBadRequest)
			return
		}

		if task.Repeat == "" {
			_, err = db.Exec("delete from scheduler where id = ?", id)
			if err != nil {
				respondJSONError(w, "Failed to delete selected task", http.StatusBadRequest)
				return
			}
		} else {
			newdate, err := repeatRule.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				respondJSONError(w, err.Error(), http.StatusBadRequest)
				return
			}

			_, err = db.Exec("UPDATE scheduler set date = ? where id = ?", newdate, id)
			if err != nil {
				respondJSONError(w, "Failed to update new data", http.StatusBadRequest)
				return
			}
		}
		respondJSON(w, JSON{}, http.StatusOK)

	default:
		return
	}
}
