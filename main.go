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

// HashPassword хэширует переданный пароль
func HashPassword(password string) (string, error) {
	// Используем bcrypt для хэширования пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// Возвращаем хэш в виде строки
	return string(hash), nil
}

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
	category := r.FormValue("category")

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

	insertQuery := "INSERT INTO testcases (name, description, project, status, type, createdby, category) VALUES ($1, $2, $3, $4, $5, $6, $7)"
	_, err = db.Exec(insertQuery, name, description, parent, status, tp, id, category)
	if err != nil {
		fmt.Println("Ошибка выполнения POST-запроса:", err)
		return
	}

	fmt.Println("POST-запрос выполнен успешно.")

	testCasePath := fmt.Sprintf("/testcases/?id=%s", parent)
	http.Redirect(w, r, testCasePath, http.StatusSeeOther)

}

func createUser(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса - только POST-запросы обрабатываются
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	session, _ := sessionStore.Get(r, "session-name")
	id := session.Values["id"]
	isAdmin := session.Values["isAdmin"].(bool)

	if id != 0 && isAdmin {
		// Получаем данные из формы
		name := r.FormValue("name")
		login := r.FormValue("login")
		isAdminform := r.FormValue("IsAdmin")
		password := r.FormValue("password")

		var isAdminformbool bool
		if isAdminform == "true" {
			isAdminformbool = true
		} else {
			isAdminformbool = false
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

		fmt.Print(isAdminform)
		passhash, _ := HashPassword(password)
		insertQuery := "INSERT INTO users (name, login, isadmin, passhash) VALUES ($1, $2, $3, $4)"
		_, err = db.Exec(insertQuery, name, login, isAdminformbool, passhash)
		if err != nil {
			fmt.Println("Ошибка выполнения POST-запроса:", err)
			return
		}

		fmt.Println("POST-запрос выполнен успешно.")

		http.Redirect(w, r, "/settings/", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
	}
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
		} else if table == "testcases" {
			testCasePath := fmt.Sprintf("/testcases/?id=%s", parent)
			http.Redirect(w, r, testCasePath, http.StatusSeeOther)
		} else if table == "users" {
			http.Redirect(w, r, "/settings/", http.StatusSeeOther)
		}
	} else {
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
	}

}

func editSubject(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	session, _ := sessionStore.Get(r, "session-name")
	id := session.Values["id"]
	isAdminSess := session.Values["isAdmin"]
	if id != 0 {
		// Получаем данные из формы
		name := r.FormValue("name")
		description := r.FormValue("description")
		tp := r.FormValue("type")
		status := r.FormValue("status")
		idsubj := r.FormValue("idsubj")
		table := r.FormValue("table")
		category := r.FormValue("category")
		rolest := r.FormValue("isadmin")
		role, _ := strconv.ParseBool(rolest)

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

		var editQuery string
		if table == "testcases" {
			editQuery = fmt.Sprintf("UPDATE %s SET name = '%s', description = '%s', type = '%s', status = '%s', category = '%s' WHERE id = $1", table, name, description, tp, status, category)
		} else if table == "projects" {
			editQuery = fmt.Sprintf("UPDATE %s SET name = '%s', description = '%s' WHERE id = $1", table, name, description)
		} else if table == "users" {
			oldpass := r.FormValue("oldpassword")
			newpass := r.FormValue("newpassword")
			if oldpass == "" {
				editQuery = fmt.Sprintf("UPDATE %s SET name = '%s', isadmin = '%t' WHERE id = $1", table, name, role)
			} else {
				rows, err := db.Query("SELECT passhash FROM users WHERE id = $1", idsubj)
				if err != nil {
					http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
					return
				}
				var passhash string
				for rows.Next() {
					if err := rows.Scan(&passhash); err != nil {
						http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
						return
					}
				}
				err = bcrypt.CompareHashAndPassword([]byte(passhash), []byte(oldpass))
				if err != nil && isAdminSess != true {
					fmt.Println("Пароль не совпал!")
					Path := fmt.Sprintf("/profile/?id=%s&passerr=1", idsubj)
					http.Redirect(w, r, Path, http.StatusSeeOther)
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				} else {
					fmt.Println("Пароль совпал")
					newpasshash, _ := HashPassword(newpass)
					editQuery = fmt.Sprintf("UPDATE %s SET name = '%s', isadmin = '%t', passhash = '%s' WHERE id = $1", table, name, role, newpasshash)
				}
			}
		}

		result, err := db.Exec(editQuery, idsubj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(w, "Изменено строк: %d", rowsAffected)
		fmt.Println("POST-запрос выполнен успешно.")

		var testCasePath string
		/*if table == "projects" {
			testCasePath = fmt.Sprintf("/testcases/?id=%s", idsubj)
			http.Redirect(w, r, testCasePath, http.StatusSeeOther)
		} else {
			testCasePath = fmt.Sprintf("/case/?id=%s", idsubj)
			http.Redirect(w, r, testCasePath, http.StatusSeeOther)
		}*/

		switch table {
		case "projects":
			testCasePath = fmt.Sprintf("/testcases/?id=%s", idsubj)
			http.Redirect(w, r, testCasePath, http.StatusSeeOther)
		case "testcases":
			testCasePath = fmt.Sprintf("/case/?id=%s", idsubj)
			http.Redirect(w, r, testCasePath, http.StatusSeeOther)
		case "users":
			testCasePath = fmt.Sprintf("/profile/?id=%s", idsubj)
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

	type PageData struct {
		Projects []Project
		IsAdmin  bool
		ID       int
	}

	// Получаем сессию
	session, _ := sessionStore.Get(r, "session-name")

	// Получаем значение из сессии
	id, _ := session.Values["id"].(int)
	isAdmin, _ := session.Values["isAdmin"].(bool)
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
		rows, err := db.Query("SELECT name, description, id FROM projects ORDER BY id")
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

		pagedata := PageData{
			Projects: projects,
			IsAdmin:  isAdmin,
			ID:       id,
		}
		tmpl, err := template.ParseFiles("templates/projects.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, pagedata); err != nil {
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
		Category    int
	}

	type CaseCount struct {
		Success string
		Failed  string
		Waiting string
	}

	type PageData struct {
		Id        int
		IsAdmin   bool
		Createdby int
		Idproj    int
		ProjName  string
		ProjDesc  string
		FnArrSt   CaseCount
		NfnArrSt  CaseCount
		RegrArrSt CaseCount
		IntArrSt  CaseCount
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
		rows, err := db.Query("SELECT name, description, id, status, type, category FROM testcases WHERE project = $1", idproj)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
		var cases []Case
		var sum int
		for rows.Next() {
			var cs Case
			if err := rows.Scan(&cs.Name, &cs.Descritpion, &cs.ID, &cs.Status, &cs.Tp, &cs.Category); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
			cases = append(cases, cs)
			sum++
		}

		var createdby int
		var projname string
		var projdesc string
		// Выполняем SQL-запрос.
		rows, err = db.Query("SELECT name, description, createdby FROM projects WHERE id = $1", idproj)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			if err := rows.Scan(&projname, &projdesc, &createdby); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
		}

		tmpl, err := template.ParseFiles("templates/testcases.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		//[Количество успешных, количество неуспешных, количество непроверенных]
		var fnArr [3]float32
		var sumFn float32
		var fnArrSt [3]string
		var nfnArr [3]float32
		var sumNfn float32
		var nfnArrSt [3]string
		var intArr [3]float32
		var sumInt float32
		var intArrSt [3]string
		var regrArr [3]float32
		var sumRegr float32
		var regrArrSt [3]string

		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 0 and status = 1 and project = $1", idproj).Scan(&fnArr[0])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 0 and status = 2 and project = $1", idproj).Scan(&fnArr[1])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 0 and status = 0 and project = $1", idproj).Scan(&fnArr[2])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 0").Scan(&sumFn)
		for i := 0; i < 3; i++ {
			fnArr[i] = fnArr[i] / sumFn * 100
			fnArrSt[i] = fmt.Sprintf("%.0f", fnArr[i])
		}
		fnArrStruct := CaseCount{
			Success: fnArrSt[0],
			Failed:  fnArrSt[1],
			Waiting: fnArrSt[2],
		}

		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 1 and status = 1 and project = $1", idproj).Scan(&nfnArr[0])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 1 and status = 2 and project = $1", idproj).Scan(&nfnArr[1])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 1 and status = 0 and project = $1", idproj).Scan(&nfnArr[2])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 1").Scan(&sumNfn)
		for i := 0; i < 3; i++ {
			nfnArr[i] = nfnArr[i] / sumNfn * 100
			nfnArrSt[i] = fmt.Sprintf("%.0f", nfnArr[i])
		}
		nfnArrStruct := CaseCount{
			Success: nfnArrSt[0],
			Failed:  nfnArrSt[1],
			Waiting: nfnArrSt[2],
		}

		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 2 and status = 1 and project = $1", idproj).Scan(&intArr[0])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 2 and status = 2 and project = $1", idproj).Scan(&intArr[1])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 2 and status = 0 and project = $1", idproj).Scan(&intArr[2])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 2").Scan(&sumInt)
		for i := 0; i < 3; i++ {
			intArr[i] = intArr[i] / sumInt * 100
			intArrSt[i] = fmt.Sprintf("%.0f", intArr[i])
		}
		intArrStruct := CaseCount{
			Success: intArrSt[0],
			Failed:  intArrSt[1],
			Waiting: intArrSt[2],
		}

		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 3 and status = 1 and project = $1", idproj).Scan(&regrArr[0])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 3 and status = 2 and project = $1", idproj).Scan(&regrArr[1])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 3 and status = 0 and project = $1", idproj).Scan(&regrArr[2])
		db.QueryRow("SELECT COUNT(*) FROM testcases WHERE category = 3").Scan(&sumRegr)
		for i := 0; i < 3; i++ {
			regrArr[i] = regrArr[i] / sumRegr * 100
			regrArrSt[i] = fmt.Sprintf("%.0f", regrArr[i])

		}
		regrArrStruct := CaseCount{
			Success: regrArrSt[0],
			Failed:  regrArrSt[1],
			Waiting: regrArrSt[2],
		}

		idprojint, _ := strconv.Atoi(idproj)
		pgData := PageData{
			Id:        id,
			IsAdmin:   isAdmin,
			Createdby: createdby,
			Idproj:    idprojint,
			CasePg:    cases,
			ProjName:  projname,
			ProjDesc:  projdesc,
			FnArrSt:   fnArrStruct,
			NfnArrSt:  nfnArrStruct,
			RegrArrSt: regrArrStruct,
			IntArrSt:  intArrStruct,
		}
		if err := tmpl.Execute(w, pgData); err != nil {
			http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
			return
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func getCase(w http.ResponseWriter, r *http.Request) {

	type PageData struct {
		Name          string
		Descritpion   string
		Descritpionbr string
		ID            int
		Status        int
		Tp            int
		IsAdmin       bool
		Iduser        int
		Createdby     int
		Project       int
		Category      int
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
		rows, err := db.Query("SELECT name, description, id, status, type, createdby, project, category FROM testcases WHERE id = $1", idcase)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}

		var cs PageData
		cs.IsAdmin = isAdmin
		cs.Iduser = id
		for rows.Next() {
			if err := rows.Scan(&cs.Name, &cs.Descritpion, &cs.ID, &cs.Status, &cs.Tp, &cs.Createdby, &cs.Project, &cs.Category); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
			cs.Descritpionbr = strings.Replace(cs.Descritpion, "\n", "<br>", -1)
			//cs.Descritpion = strings.Replace(cs.Descritpion, "\n", "&#13;&#10", -1)
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

func getProfile(w http.ResponseWriter, r *http.Request) {

	type PageData struct {
		Name        string
		Login       string
		ID          int
		IsAdmin     bool
		Iduser      int
		IsAdminsess bool
		PassErr     bool
	}

	// Получаем сессию
	session, _ := sessionStore.Get(r, "session-name")

	// Получаем значение из сессии
	id, _ := session.Values["id"].(int)
	isAdmin, _ := session.Values["isAdmin"].(bool)
	if id != 0 {

		// Получение значения параметра "id" из URL
		iduser := r.URL.Query().Get("id")
		passErr := r.URL.Query().Get("passerr")
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
		rows, err := db.Query("SELECT name, login, isadmin FROM users WHERE id = $1", iduser)
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}

		var user PageData
		user.IsAdminsess = isAdmin
		user.ID = id
		user.Iduser, _ = strconv.Atoi(iduser)
		user.PassErr, _ = strconv.ParseBool(passErr)
		for rows.Next() {
			if err := rows.Scan(&user.Name, &user.Login, &user.IsAdmin); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
		}
		tmpl, err := template.ParseFiles("templates/profile.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		if err := tmpl.Execute(w, user); err != nil {
			http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
			return
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {

	type User struct {
		Name  string
		Login string
		ID    int
		Admin bool
	}

	type PageData struct {
		Users  []User
		IdUser int
	}

	// Получаем сессию
	session, _ := sessionStore.Get(r, "session-name")

	// Получаем значение из сессии
	id, _ := session.Values["id"].(int)
	isAdmin, _ := session.Values["isAdmin"].(bool)

	if id != 0 && isAdmin {

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
		rows, err := db.Query("SELECT name, login, id, isadmin FROM users")
		if err != nil {
			http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
			return
		}
		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.Name, &user.Login, &user.ID, &user.Admin); err != nil {
				http.Error(w, "Ошибка сканирования строк", http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		tmpl, err := template.ParseFiles("templates/settings.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки HTML-шаблона", http.StatusInternalServerError)
			return
		}

		pgData := PageData{
			Users:  users,
			IdUser: id,
		}
		if err := tmpl.Execute(w, pgData); err != nil {
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
		session, _ := sessionStore.Get(r, "session-name")
		idsess := session.Values["id"]
		if idsess != 0 {
			w.Header().Set("Content-Type", "text/html")
			http.Redirect(w, r, "/projects", http.StatusSeeOther)
		} else {
			w.Header().Set("Content-Type", "text/html")
			http.ServeFile(w, r, "templates/login.html")
		}
	})

	http.HandleFunc("/projects", getProjects)

	http.HandleFunc("/testcases/", getCases)

	http.HandleFunc("/case/", getCase)

	http.HandleFunc("/check", formAuth)

	http.HandleFunc("/createproject", formCreateHandler)

	http.HandleFunc("/testcases/createcase", caseCreateHandler)

	http.HandleFunc("/testcases/deletesubject", deleteSubject)

	http.HandleFunc("/case/deletesubject", deleteSubject)

	http.HandleFunc("/case/editsubject", editSubject)

	http.HandleFunc("/testcases/editsubject", editSubject)

	http.HandleFunc("/profile/", getProfile)

	http.HandleFunc("/profile/editsubject", editSubject)

	http.HandleFunc("/profile/deletesubject", deleteSubject)

	http.HandleFunc("/settings/", getUsers)

	http.HandleFunc("/settings/createuser", createUser)

	http.HandleFunc("/logout", logoutHandler)

	// Запускаем веб-сервер на порту 8080.
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
