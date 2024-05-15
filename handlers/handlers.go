package handlers

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "net/http"
    "time"
    "errors"

    _ "modernc.org/sqlite"

    "github.com/noctusha/finalya/repeatRule"
    "github.com/noctusha/finalya/connection"
)

type Task struct {
    Date    string  `json:"date"`
    Title   string  `json:"title"`
    Comment string  `json:"comment"`
    Repeat  string  `json:"repeat"`
}

type JSON struct {
    ID int64 `json:"id,omitempty"`
    Err error `json:"error,omitempty"`
    Tasks *[]Task `json:"tasks,omitempty"`
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
    now := r.URL.Query().Get("now")
    date := r.URL.Query().Get("date")
    repeat := r.URL.Query().Get("repeat")

    // Проверка наличия параметров
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
    respondJSON(w, JSON{Err: errors.New(message)}, statusCode)
}


func NewTaskHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
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

    default:
        return
    }
}

func ListTasksHandler(w http.ResponseWriter, r *http.Request) {
        tasks := []Task{}
        task := Task{}
        db := connection.ConnectingDB()
        defer db.Close()


            rows, err := db.Query("select date, title, comment, repeat from scheduler order by date limit :limit", sql.Named("limit", 20))

            if err != nil {
                respondJSONError(w, "Failed to select task from database", http.StatusBadRequest)
                return
            }
            defer rows.Close()

            for rows.Next() {
                err := rows.Scan(&task.Date, &task.Title, &task.Comment, &task.Repeat)
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
