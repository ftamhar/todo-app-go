package datastore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"

	"github.com/Xanvial/todo-app-go/model"
	_ "github.com/lib/pq"
)

type DBStore struct {
	db *sql.DB
}

func NewDBStore() *DBStore {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		model.DBHost, model.DBPort, model.DBUser, model.DBPassword, model.DBName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB Successfully connected!")

	return &DBStore{
		db: db,
	}
}

func (ds *DBStore) GetCompleted(w http.ResponseWriter, r *http.Request) {
	var completed []model.TodoData

	query := `
		SELECT id, title, status
		FROM todo
		WHERE status = true
	`

	rows, err := ds.db.Query(query)
	if err != nil {
		log.Println("error on getting todo:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	defer rows.Close()

	for rows.Next() {
		var data model.TodoData
		if err := rows.Scan(&data.ID, &data.Title, &data.Status); err != nil {
			log.Println("error on getting todo:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		completed = append(completed, data)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completed)
}

func (ds *DBStore) GetIncomplete(w http.ResponseWriter, r *http.Request) {
	var incompleted []model.TodoData

	query := `
		SELECT id, title, status
		FROM todo
		WHERE status = false
	`

	rows, err := ds.db.Query(query)
	if err != nil {
		log.Println("error on getting todo:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	defer rows.Close()

	for rows.Next() {
		var data model.TodoData
		if err := rows.Scan(&data.ID, &data.Title, &data.Status); err != nil {
			log.Println("error on getting todo:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		incompleted = append(incompleted, data)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incompleted)
}

func (ds *DBStore) CreateTodo(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	query := `
		INSERT INTO todo(title) values ($1) RETURNING id
`
	id := 0
	ds.db.QueryRow(query, title).Scan(&id)
	res := model.TodoData{
		ID:     id,
		Title:  title,
		Status: false,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (ds *DBStore) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	value := r.FormValue("status")

	w.Header().Set("Content-Type", "application/json")
	status, err := strconv.ParseBool(value)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "status must be boolean",
		})
		return
	}
	idx, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "id must be numeric",
		})
		return
	}
	q := `
		UPDATE todo SET status = $1 WHERE id = $2 RETURNING id, title, status
`
	var res model.TodoData
	err = ds.db.QueryRow(q, status, idx).Scan(&res.ID, &res.Title, &res.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (ds *DBStore) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["id"]
	idx, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "id must be numeric",
		})
		return
	}

	q := `
		DELETE FROM todo WHERE id = $1
`
	_, err = ds.db.Exec(q, idx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
