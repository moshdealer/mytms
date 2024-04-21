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

// Обработчик, который будет доступен только после аутентификации.
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

// User структура представляет собой модель пользователя.
type User struct {
	Username string
	Password []byte
}

// временная "База данных" пользователей.
var users = map[string]User{
	"user1": {Username: "user1", Password: hashPassword("password1")},
	"user2": {Username: "user2", Password: hashPassword("password2")},
}

// Функция для хэширования пароля.
func hashPassword(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}

// Функция для проверки пароля пользователя.
func verifyPassword(user User, password string) bool {
	err := bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	return err == nil
}

// Обработчик, который требует аутентификации.
func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || !authenticate(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized\n"))
			return
		}
		handler(w, r)
	}
}

// Обработчик, который будет доступен только после аутентификации.
func restrictedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the restricted area, %s!\n", r.RemoteAddr)
}

// Функция для проверки аутентификации.
func authenticate(username, password string) bool {
	user, found := users[username]
	if !found {
		return false
	}
	return verifyPassword(user, password)
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
	rows, err := db.Query("SELECT name, description FROM projects")
	if err != nil {
		http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
		return
	}
	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.Name, &project.Descritpion); err != nil {
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

	http.HandleFunc("/check", formHandler)

	http.HandleFunc("/createproject", formCreateHandler)

	// Запускаем веб-сервер на порту 8080.
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
