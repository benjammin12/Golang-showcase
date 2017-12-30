package main

import (
	"net/http"
	"fmt"
	"prog-assignment-golang/session"
	"html/template"
	"os"
	"golang.org/x/crypto/bcrypt"
	"github.com/satori/go.uuid"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"log"
	"github.com/gorilla/mux"
)

var (
	globalSessions *session.Manager
	err error
	Templ *template.Template
	Db *sql.DB
)

func init(){
	//globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	//go globalSessions.GC()
	connectionName := mustGetenv("CLOUDSQL_CONNECTION_NAME")
	userName := mustGetenv("CLOUDSQL_USER")
	password := os.Getenv("CLOUDSQL_PASSWORD")


	Db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@cloudsql(%s)/Tasks", userName, password, connectionName))
	if err != nil {
		fmt.Println("Connection failed")
		fmt.Println(err.Error())
		return
	}else {
		fmt.Println("Connection okay")
	}
	Templ = template.Must(template.ParseGlob("templates/*"))


	r := mux.NewRouter()
	r.PathPrefix("/style").Handler(http.StripPrefix("/style/",http.FileServer(http.Dir("style"))))
	r.PathPrefix("/public").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	r.HandleFunc("/", index)
	r.HandleFunc("/login", loginPage).Methods("GET")
	r.HandleFunc("/login", loginUser).Methods("POST")
	r.HandleFunc("/seed", addUser)

	//used for GAE to pick up routes
	http.Handle("/", r)

	/* USER
	+----------+--------------+------+-----+---------+----------------+
	| Field    | Type         | Null | Key | Default | Extra          |
	+----------+--------------+------+-----+---------+----------------+
	| id       | int(11)      | NO   | PRI | NULL    | auto_increment |
	| name     | varchar(255) | NO   |     | NULL    |                |
	| password | varchar(255) | NO   |     | NULL    |                |
	+----------+--------------+------+-----+---------+----------------+
	*/

	/*TASKS
	+-----------+--------------+------+-----+---------+----------------+
	| Field     | Type         | Null | Key | Default | Extra          |
	+-----------+--------------+------+-----+---------+----------------+
	| id        | int(11)      | NO   | PRI | NULL    | auto_increment |
	| task_name | varchar(255) | NO   |     | NULL    |                |
	| task_desc | text         | YES  |     | NULL    |                |
	+-----------+--------------+------+-----+---------+----------------+
	 */


}
func CheckLoginStatus(w http.ResponseWriter, r *http.Request) (bool){
	sess := globalSessions.SessionStart(w,r)
	sess_uid := sess.Get("UserID")
	//u := model.MainUser{}
	if sess_uid == nil {
		//http.Redirect(w,r, "/", http.StatusForbidden)
		//Tpl.ExecuteTemplate(w,"index", "You can't access this page")
		return false
	} else {
		uID := sess_uid
		fmt.Println("Logged in User, ", uID)
		//Tpl.ExecuteTemplate(w, "user", nil)
		return true
	}
}

func index(w http.ResponseWriter, r *http.Request){
	err := Templ.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func loginPage(w http.ResponseWriter, r *http.Request){
	err := Templ.ExecuteTemplate(w, "login", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func loginUser(w http.ResponseWriter, r *http.Request){
	sess := globalSessions.SessionStart(w, r)

	if r.Method != "POST" {
		http.ServeFile(w,r, "login.html")
		return
	}

	username := r.FormValue("name")
	password := r.FormValue("password")

	var databaseUserName, databasePassword string

	err := Db.QueryRow("SELECT name,password FROM user WHERE name=?", username).Scan(&databaseUserName, &databasePassword)
	//no user found
	if err != nil {
		Templ.ExecuteTemplate(w, "login" ,"Username and Password did not match! Please try again")
		return
	}

	//wrong password
	if err := bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password)); err != nil {
		//log.Fatal("Error comparing passwords", err)
		Templ.ExecuteTemplate(w, "login" ,"Username and password did not match! Please try again")
		return
	} else { //Login was sucessful, create session and cookie
		u1 := uuid.NewV4() //random uuid
		sess.Set("username", r.Form["username"])
		sess.Set("UserID", u1)
		Templ.ExecuteTemplate(w, "adminHome", nil)
		return
	}
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Panicf("%s environment variable not set.", k)
	}
	return v
}

func addUser(w http.ResponseWriter, r *http.Request) {
	userName := "cull@example.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("makethefuture"), bcrypt.DefaultCost)
	if err != nil {
		// TODO: Properly handle error
		log.Fatal(err)
	}

	_, err = Db.Exec("INSERT INTO admin (name,password) values ($1,$2)",userName,hash)

	if err != nil {
		log.Fatalln(err)
	}

}