package main

import (
	"net/http"
	"fmt"

	"html/template"
	"os"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
	"database/sql"
	"prog-assignment-golang-heroku/session"
	_ "prog-assignment-golang-heroku/memory"
	"log"
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

var (
	globalSessions *session.Manager
	err error
	templ *template.Template
	db *sql.DB
)

func init(){
	globalSessions, _ = session.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()


	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}else {
		fmt.Println("Connected to Db")
	}

}


func main(){

	templ = template.Must(template.ParseGlob("templates/*"))


	r := mux.NewRouter()
	r.PathPrefix("/style").Handler(http.StripPrefix("/style/",http.FileServer(http.Dir("style"))))
	r.PathPrefix("/public").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("public"))))
	r.HandleFunc("/", index)
	r.HandleFunc("/login", loginPage).Methods("GET")
	r.HandleFunc("/login", loginUser).Methods("POST")
	r.HandleFunc("/user", userPage).Methods("GET")
	r.HandleFunc("/tasks", GetTasks).Methods("GET")
	r.HandleFunc("/tasks", AddTask).Methods("POST")



	//r.HandleFunc("/seed", addUser)

	http.ListenAndServe(":" + os.Getenv("PORT"),handlers.LoggingHandler(os.Stdout,r))

	//used for GAE to pick up routes


}
func CheckLoginStatus(w http.ResponseWriter, r *http.Request) (bool,interface{}){
	sess := globalSessions.SessionStart(w,r)
	sess_uid := sess.Get("UserID")
	if sess_uid == nil {
		http.Redirect(w,r, "/", http.StatusForbidden)
		//Tpl.ExecuteTemplate(w,"index", "You can't access this page")
		return false,""
	} else {
		uID := sess_uid
		name := sess.Get("username")
		fmt.Println("Logged in User, ", uID)
		//Tpl.ExecuteTemplate(w, "user", nil)
		return true,name
	}
}

func index(w http.ResponseWriter, r *http.Request){
	err := templ.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func loginPage(w http.ResponseWriter, r *http.Request){
	err := templ.ExecuteTemplate(w, "login", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func userPage(w http.ResponseWriter, r *http.Request){
	isLogged, name := CheckLoginStatus(w,r)

	if isLogged {
		err := templ.ExecuteTemplate(w, "userHome", name)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.Redirect(w,r,"/",http.StatusUnauthorized)
	}
}

func loginUser(w http.ResponseWriter, r *http.Request){
	sess := globalSessions.SessionStart(w, r)

	if r.Method != "POST" {
		http.ServeFile(w,r, "login.html")
		return
	}

	username := r.FormValue("email")
	password := r.FormValue("password")



	var user_id int
	var databaseUserName, databasePassword string

	row := db.QueryRow("SELECT * FROM main_user WHERE user_email = $1 ", username).Scan(&user_id,&databaseUserName, &databasePassword)
	//no user found
	if row != nil {
		templ.ExecuteTemplate(w, "login" ,"No user in db")
		return
	}

	//wrong password
	if err := bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password)); err != nil {
		//log.Fatal("Error comparing passwords", err)
		templ.ExecuteTemplate(w, "login" ,"Username and password did not match! Please try again")
		return
	} else { //Login was sucessful, create session and cookie
		u1 := user_id
		sess.Set("username", r.Form["email"])
		sess.Set("UserID", u1)
		http.Redirect(w,r, "/user", http.StatusSeeOther)
		//templ.ExecuteTemplate(w, "userHome", "Welcome " + databaseUserName)
		return
	}
}



func addUser(w http.ResponseWriter, r *http.Request) {
	userName := "cull@example.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("makethefuture"), bcrypt.DefaultCost)
	if err != nil {
		// TODO: Properly handle error
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO main_user (user_email,user_password) values ($1,$2)",userName,hash)

	if err != nil {
		log.Fatalln(err)
	}

}