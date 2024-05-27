package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const connStr = "user=postgres dbname=mytms password=password host=localhost sslmode=disable"

var sessionStore = sessions.NewCookieStore([]byte("sessionpassword"))

func restrictSlash(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "session-name")
	id := session.Values["id"]
	if id != 0 {
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func formAuth(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса - только POST-запросы обрабатываются
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	log := r.FormValue("login")
	pass := r.FormValue("pass")

	fmt.Println(log, pass)
	// Открываем соединение с базой данных.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		//log.Fatal(err)
	}
	defer db.Close()

	// Проверяем соединение с базой данных.
	err = db.Ping()
	if err != nil {
		//log.Fatal("Could not ping database:", err)
	}

	// Выполняем SQL-запрос.
	rows, err := db.Query("SELECT login, passhash, id, isadmin FROM users WHERE login = $1", log)
	if err != nil {
		//log.Fatal(err)
	}
	defer rows.Close()
	var login string
	var passhash string
	var id int
	var isAdmin bool

	// Читаем результаты запроса.
	for rows.Next() {
		err := rows.Scan(&login, &passhash, &id, &isAdmin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	// Проверяем пароль пользователя
	err = bcrypt.CompareHashAndPassword([]byte(passhash), []byte(pass))
	if err != nil {
		fmt.Println("Пароль неверен!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		fmt.Println("Пароль совпал")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Получаем сессию
		session, _ := sessionStore.Get(r, "session-name")
		// Устанавливаем значение в сессию
		session.Values["id"] = id
		session.Values["isAdmin"] = isAdmin
		session.Save(r, w)
		fmt.Print(w, "Сессия установлена")

		http.Redirect(w, r, "/projects", http.StatusSeeOther)
	}

}

func formCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса - только POST-запросы обрабатываются
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	session, _ := sessionStore.Get(r, "session-name")
	id := session.Values["id"]
	// Получаем данные из формы
	name := r.FormValue("name")
	description := r.FormValue("description")

	fmt.Println(name, description)
	// Открываем соединение с базой данных.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверяем соединение с базой данных.
	err = db.Ping()
	if err != nil {
		log.Fatal("Could not ping database:", err)
	}

	insertQuery := "INSERT INTO projects (name, description, createdby) VALUES ($1, $2, $3)"
	_, err = db.Exec(insertQuery, name, description, id)
	if err != nil {
		fmt.Println("Ошибка выполнения POST-запроса:", err)
		return
	}

	fmt.Println("POST-запрос выполнен успешно.")
	http.Redirect(w, r, "/projects", http.StatusSeeOther)

}

func caseCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса - только POST-запросы обрабатываются
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	session, _ := sessionStore.Get(r, "session-name")
	id := session.Values["id"]
	// Получаем данные из формы
	name := r.FormValue("name")
	description := r.FormValue("description")
	tp := r.FormValue("type")
	status := r.FormValue("status")
	parent := r.FormValue("parent")

	fmt.Println(name, description)
	// Открываем соединение с базой данных.
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверяем соединение с базой данных.
	err = db.Ping()
	if err != nil {
		log.Fatal("Could not ping database:", err)
	}

	insertQuery := "INSERT INTO testcases (name, description, project, status, type, createdby) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err = db.Exec(insertQuery, name, description, parent, status, tp, id)
	if err != nil {
		fmt.Println("Ошибка выполнения POST-запроса:", err)
		return
	}

	fmt.Println("POST-запрос выполнен успешно.")

	testCasePath := fmt.Sprintf("/testcases/?id=%s", parent)
	http.Redirect(w, r, testCasePath, http.StatusSeeOther)

}

func deleteSubject(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	session, _ := sessionStore.Get(r, "session-name")
	id := session.Values["id"]

	if id != 0 {
		// Получаем данные из формы
		idsubj := r.FormValue("idsubj")
		table := r.FormValue("table")
		parent := r.FormValue("parent")

		fmt.Println(idsubj, table)
		// Открываем соединение с базой данных.
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Проверяем соединение с базой данных.
		err = db.Ping()
		if err != nil {
			log.Fatal("Could not ping database:", err)
		}

		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE id = $1", table)
		fmt.Print(deleteQuery)
		result, err := db.Exec(deleteQuery, idsubj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(w, "Удалено строк: %d", rowsAffected)
		fmt.Println("POST-запрос выполнен успешно.")

		if table == "projects" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			testCasePath := fmt.Sprintf("/testcases/?id=%s", parent)
			http.Redirect(w, r, testCasePath, http.StatusSeeOther)
		}
	} else {
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
	}

}

func getProjects(w http.ResponseWriter, r *http.Request) {

	type Project struct {
		Name        string
		Descritpion string
		ID          int
	}

	// Получаем сессию
	session, _ := sessionStore.Get(r, "session-name")

	// Получаем значение из сессии
	id, _ := session.Values["id"].(int)
	if id != 0 {
		// Открываем соединение с базой данных.
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Проверяем соединение с базой данных.
		err = db.Ping()
		if err != nil {
			log.Fatal("Could not ping database:", err)
		}

		// Выполняем SQL-запрос.
		rows, err := db.Query("SELECT name, description, id FROM projects")
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
		var projects []Project
		for rows.Next() {
			var project Project
			if err := rows.Scan(&project.Name, &project.Descritpion, &project.ID); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
			projects = append(projects, project)
		}

		tmpl, err := template.ParseFiles("templates/projects.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, projects); err != nil {
			http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
			return
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func getCases(w http.ResponseWriter, r *http.Request) {

	type Case struct {
		Name        string
		Descritpion string
		ID          int
		Status      int
		Tp          int
	}

	type PageData struct {
		Id        int
		IsAdmin   bool
		Createdby int
		Idproj    int
		CasePg    []Case
	}

	// Получаем сессию
	session, _ := sessionStore.Get(r, "session-name")

	// Получаем значение из сессии
	id, _ := session.Values["id"].(int)
	isAdmin, _ := session.Values["isAdmin"].(bool)

	if id != 0 {
		// Получение значения параметра "id" из URL
		idproj := r.URL.Query().Get("id")

		// Открываем соединение с базой данных.
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Проверяем соединение с базой данных.
		err = db.Ping()
		if err != nil {
			log.Fatal("Could not ping database:", err)
		}

		// Выполняем SQL-запрос.
		rows, err := db.Query("SELECT name, description, id, status, type FROM testcases WHERE project = $1", idproj)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
		var cases []Case
		for rows.Next() {
			var cs Case
			if err := rows.Scan(&cs.Name, &cs.Descritpion, &cs.ID, &cs.Status, &cs.Tp); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
			cases = append(cases, cs)
		}

		var createdby int
		// Выполняем SQL-запрос.
		rows, err = db.Query("SELECT createdby FROM projects WHERE id = $1", idproj)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			if err := rows.Scan(&createdby); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
		}

		tmpl, err := template.ParseFiles("templates/testcases.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		idprojint, _ := strconv.Atoi(idproj)
		pgData := PageData{
			Id:        id,
			IsAdmin:   isAdmin,
			Createdby: createdby,
			Idproj:    idprojint,
			CasePg:    cases,
		}
		if err := tmpl.Execute(w, pgData); err != nil {
			http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
			return
		}
		fmt.Print(session.Options.MaxAge)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func getCase(w http.ResponseWriter, r *http.Request) {

	type PageData struct {
		Name        string
		Descritpion string
		ID          int
		Status      int
		Tp          int
		IsAdmin     bool
		Iduser      int
		Createdby   int
		Project     int
	}

	// Получаем сессию
	session, _ := sessionStore.Get(r, "session-name")

	// Получаем значение из сессии
	id, _ := session.Values["id"].(int)
	isAdmin, _ := session.Values["isAdmin"].(bool)
	if id != 0 {

		// Получение значения параметра "id" из URL
		idcase := r.URL.Query().Get("id")
		// Открываем соединение с базой данных.
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Проверяем соединение с базой данных.
		err = db.Ping()
		if err != nil {
			log.Fatal("Could not ping database:", err)
		}

		// Выполняем SQL-запрос.
		rows, err := db.Query("SELECT name, description, id, status, type, createdby, project FROM testcases WHERE id = $1", idcase)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}

		var cs PageData
		cs.IsAdmin = isAdmin
		cs.Iduser = id
		for rows.Next() {
			if err := rows.Scan(&cs.Name, &cs.Descritpion, &cs.ID, &cs.Status, &cs.Tp, &cs.Createdby, &cs.Project); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
			cs.Descritpion = strings.Replace(cs.Descritpion, "\n", "<br>", -1)
		}

		tmpl, err := template.ParseFiles("templates/case.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, cs); err != nil {
			http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
			return
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, "session-name")

	// Удаление сессии путем установки отрицательного времени жизни
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func main() {
	// Устанавливаем обработчик для маршрута /restricted, который требует аутентификации.
	http.HandleFunc("/", restrictSlash)

	// Настройка обработчика для статических файлов.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Обработчик запроса на SVG-файлы.
	http.HandleFunc("/static/images/", func(w http.ResponseWriter, r *http.Request) {
		// Установка MIME-типа для SVG.
		w.Header().Set("Content-Type", "image/svg+xml")

		// Обслуживание файла SVG.
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("static/styles/"))))

	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем правильный MIME-тип для файлов SVG.
		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "templates/login.html")
	})

	http.HandleFunc("/projects", getProjects)

	http.HandleFunc("/testcases/", getCases)

	http.HandleFunc("/case/", getCase)

	http.HandleFunc("/check", formAuth)

	http.HandleFunc("/createproject", formCreateHandler)

	http.HandleFunc("/testcases/createcase", caseCreateHandler)

	http.HandleFunc("/testcases/deletesubject", deleteSubject)

	http.HandleFunc("/case/deletesubject", deleteSubject)

	http.HandleFunc("/logout", logoutHandler)

	// Запускаем веб-сервер на порту 8080.
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
