package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"log"

	"github.com/noctusha/finalya/connection"
	"github.com/noctusha/finalya/models"
	"github.com/noctusha/finalya/repeatRule"
)

type Handler struct {
	Repo *connection.Repository
}

type JSON struct {
	ID    int64          `json:"id,omitempty"`
	Err   string         `json:"error,omitempty"`
	Tasks *[]models.Task `json:"tasks,omitempty"`
}

func respondJSON(w http.ResponseWriter, payload interface{}, statusCode int) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte("Internal server error"))
		if writeErr != nil {
			log.Printf("Error writing an error in respondJSON: %v", writeErr)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(response)
		if writeErr != nil {
			log.Printf("Error writing response in respondJSON: %v", writeErr)
		}
}

func respondJSONError(w http.ResponseWriter, message string, statusCode int) {
	respondJSON(w, JSON{Err: message}, statusCode)
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
	_, writeErr := w.Write(([]byte(nextDate)))
		if writeErr != nil {
			log.Printf("Error writing response in NextDateHandler: %v", writeErr)
		}
}

func (h *Handler) NewTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var task models.Task

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

	id, err := h.Repo.AddTask(task)
	if err != nil {
		respondJSONError(w, "Failed to insert task into database", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{ID: id}, http.StatusOK)
}

func (h *Handler) ChangeTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	task, err := h.Repo.GetTaskByID(id)
	if err != nil {
		respondJSONError(w, "Failed to scan selected result from database", http.StatusBadRequest)
		return
	}

	respondJSON(w, task, http.StatusOK)
}

func (h *Handler) PushChangedTask(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var task models.Task

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

	err = h.Repo.UpdateTask(task)
	if err != nil {
		respondJSONError(w, "Failed to update new data", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{}, http.StatusOK)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	err := h.Repo.DeleteTask(id)
	if err != nil {
		respondJSONError(w, "Failed to delete selected task", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{}, http.StatusOK)
}

func (h *Handler) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.NewTask(w, r)

	case http.MethodGet:
		h.ChangeTask(w, r)

	case http.MethodPut:
		h.PushChangedTask(w, r)

	case http.MethodDelete:
		h.DeleteTask(w, r)

	default:
		return
	}
}

func (h *Handler) ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	tasks, err := h.Repo.ListTasks(20, search)
	if err != nil {
		respondJSONError(w, "Failed to select task from database", http.StatusBadRequest)
		return
	}

	respondJSON(w, JSON{Tasks: &tasks}, http.StatusOK)
}

func (h *Handler) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		id := r.URL.Query().Get("id")

		task, err := h.Repo.GetTaskByID(id)
		if err != nil {
			respondJSONError(w, "Failed to scan selected result from database", http.StatusBadRequest)
			return
		}

		if task.Repeat == "" {
			err = h.Repo.DeleteTask(id)
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

			task.Date = newdate
			err = h.Repo.UpdateTask(task)
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
