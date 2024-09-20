package rest

import (
	"database/sql"
	"encoding/json"
	"log"
	"mytms/internal/config"
	"net/http"

	_ "github.com/lib/pq"
)

var cfg = config.LoadConfig()
var connStr = config.MakeBDPath(*cfg)

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Createdby   string `json:"createdby"`
}

type TestCase struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Project     string `json:"project"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	Createdby   string `json:"createdby"`
	Category    string `json:"category"`
}

func GetProjectsREST(w http.ResponseWriter, r *http.Request) {
	idproj := r.URL.Query().Get("id")

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("БД недоступна: ", err)
	}

	var rows *sql.Rows

	if idproj == "" {
		// Выполняем SQL-запрос без параметра
		rows, err = db.Query("SELECT id, name, description, createdby FROM projects ORDER BY id")
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
	} else {
		// Выполняем SQL-запрос с параметром
		rows, err = db.Query("SELECT id, name, description, createdby FROM projects where id = $1", idproj)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
	}

	var projects []Project
	for rows.Next() {
		var project Project
		err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.Createdby)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		projects = append(projects, project)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func GetTestCasesREST(w http.ResponseWriter, r *http.Request) {
	idcase := r.URL.Query().Get("id")

	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("БД недоступна:", err)
	}

	var rows *sql.Rows
	if idcase == "" {
		// Выполняем SQL-запрос без параметра
		rows, err = db.Query("SELECT id, name, description, project, status, type, createdby, category FROM testcases ORDER BY id")
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
	} else {
		// Выполняем SQL-запрос с параметром
		rows, err = db.Query("SELECT id, name, description, project, status, type, createdby, category FROM testcases WHERE id = $1", idcase)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
	}

	var testcases []TestCase
	for rows.Next() {
		var testcase TestCase
		err := rows.Scan(&testcase.ID, &testcase.Name, &testcase.Description, &testcase.Project, &testcase.Status, &testcase.Type, &testcase.Createdby, &testcase.Category)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		testcases = append(testcases, testcase)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testcases)
}

func PushProject(w http.ResponseWriter, r *http.Request) {
	var proj Project

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("БД недоступна:", err)
	}

	// Декодирование JSON-запроса в структуру
	if err := json.NewDecoder(r.Body).Decode(&proj); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO projects (name, description, createdby) VALUES ($1, $2, $3) RETURNING id", proj.Name, proj.Description, proj.Createdby).Scan(&proj.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(proj)
}

func PushCase(w http.ResponseWriter, r *http.Request) {
	var testcase TestCase

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("БД недоступна:", err)
	}

	// Декодирование JSON-запроса в структуру
	if err := json.NewDecoder(r.Body).Decode(&testcase); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO testcases (name, description, project, status, type, createdby, category) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", testcase.Name, testcase.Description, testcase.Project, testcase.Status, testcase.Type, testcase.Createdby, testcase.Category).Scan(&testcase.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(testcase)
}
