package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const connStr = "user=postgres dbname=mytms password=password host=localhost sslmode=disable"
const sessionPass = "password"

func restrictSlash(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte(sessionPass))
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	val := session.Values["isAuth"]
	fmt.Println(val)
	if val == true {
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {
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
	rows, err := db.Query("SELECT login, passhash FROM users WHERE login = $1", log)
	if err != nil {
		//log.Fatal(err)
	}
	defer rows.Close()

	var login string
	var passhash string

	// Читаем результаты запроса.
	for rows.Next() {
		err := rows.Scan(&login, &passhash)
		if err != nil {
			//log.Fatal(err)
		}
		fmt.Print("new", login, passhash)
	}

	// Проверяем пароль пользователя
	err = bcrypt.CompareHashAndPassword([]byte(passhash), []byte(pass))
	if err != nil {
		fmt.Println("Пароль неверен!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		fmt.Println("Пароль совпал")
		var store = sessions.NewCookieStore([]byte("4n0570JlM4I2ruH4L"))
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Получаем isAuth значение из сессии
		val := session.Values["isAuth"]
		if val == nil {
			// Если значения нет, устанавливаем его
			session.Values["isAuth"] = true
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println("Значение isAuth установлено в сессию")
		} else {
			fmt.Println("Значение isAuth из сессии: ", val)
		}

		// Получаем log значение из сессии
		val = session.Values["log"]
		if val == nil {
			// Если значения нет, устанавливаем его
			session.Values["log"] = log
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println("Значение log установлено в сессию")
		} else {
			fmt.Println("Значение log из сессии: ", val.(string))
		}
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
	}

}

func formCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса - только POST-запросы обрабатываются
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

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

	insertQuery := "INSERT INTO projects (name, description) VALUES ($1, $2)"
	_, err = db.Exec(insertQuery, name, description)
	if err != nil {
		fmt.Println("Ошибка выполнения POST-запроса:", err)
		return
	}

	fmt.Println("POST-запрос выполнен успешно.")
	http.Redirect(w, r, "/projects", http.StatusSeeOther)

}

func getProjects(w http.ResponseWriter, r *http.Request) {

	type Project struct {
		Name        string
		Descritpion string
		ID          int
	}

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

}

func getCases(w http.ResponseWriter, r *http.Request) {

	type Case struct {
		Name        string
		Descritpion string
		ID          int
		Status      int
		Tp          int
	}

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

	tmpl, err := template.ParseFiles("templates/testcases.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, cases); err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		return
	}
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

	http.HandleFunc("/check", formHandler)

	http.HandleFunc("/createproject", formCreateHandler)

	// Запускаем веб-сервер на порту 8080.
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
