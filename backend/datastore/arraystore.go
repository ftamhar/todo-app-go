package datastore

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"

	"github.com/Xanvial/todo-app-go/model"
)

type ArrayStore struct {
	data []model.TodoData
}

func NewArrayStore() *ArrayStore {
	newData := make([]model.TodoData, 0)

	return &ArrayStore{
		data: newData,
	}
}

func (as *ArrayStore) GetCompleted(w http.ResponseWriter, r *http.Request) {
	// get completed data
	completed := make([]model.TodoData, 0)
	for _, d := range as.data {
		if d.Status {
			completed = append(completed, d)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completed)
}

func (as *ArrayStore) GetIncomplete(w http.ResponseWriter, r *http.Request) {
	incomplete := make([]model.TodoData, 0)

	for i := range as.data {
		if !as.data[i].Status {
			incomplete = append(incomplete, as.data[i])
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incomplete)
}

func (as *ArrayStore) CreateTodo(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")

	newTodo := model.TodoData{
		ID:     len(as.data) + 1,
		Title:  title,
		Status: false,
	}
	as.data = append(as.data, newTodo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newTodo)
}

func (as *ArrayStore) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	status := r.FormValue("status")
	idx, _ := strconv.Atoi(id)

	var res model.TodoData
	for i := range as.data {
		if as.data[i].ID == idx {
			parseBool, err := strconv.ParseBool(status)
			if err != nil {
				w.WriteHeader(http.StatusNotAcceptable)
				w.Write([]byte("failed to parse bool"))
				return
			}
			as.data[i].Status = parseBool
			res = as.data[i]
			break
		}
	}

	if res.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("data not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (as *ArrayStore) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	w.Header().Set("Content-Type", "application/json")
	idx, err := strconv.Atoi(id)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "id is not numeric",
		})
		return
	}

	delId := -1
	for i := range as.data {
		if as.data[i].ID == idx {
			delId = i
			break
		}
	}

	as.data = append(as.data[:delId], as.data[delId+1:]...)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
