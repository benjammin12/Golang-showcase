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

	//check if the user is authenticated before adding/viewing tasks
	r.Handle("/tasks", isAuthenticated(GetTasks)).Methods("GET")
	r.Handle("/tasks", isAuthenticated(AddTask)).Methods("POST")

	r.HandleFunc("/unauthorized", unauthorized)


	http.ListenAndServe(":" + os.Getenv("PORT"),handlers.LoggingHandler(os.Stdout,r))

}


/*Helper function to check if the user is authenticated
 * Returns an bool representing if the person is logged in and an interface representing their name
 */
func CheckLoginStatus(w http.ResponseWriter, r *http.Request) (bool,interface{}){
	sess := globalSessions.SessionStart(w,r)
	sess_uid := sess.Get("UserID")
	if sess_uid == nil {
		return false,""
	} else {
		uID := sess_uid
		name := sess.Get("username")
		fmt.Println("Logged in User, ", uID)
		//Tpl.ExecuteTemplate(w, "user", nil)
		return true,name
	}
}

/* Routes to home page
 *
 */
func index(w http.ResponseWriter, r *http.Request){
	err := templ.ExecuteTemplate(w, "index", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

/* Routes to Login Page
 *
 */
func loginPage(w http.ResponseWriter, r *http.Request){
	err := templ.ExecuteTemplate(w, "login", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

/* Routes to unauthorized page if the user attempts to access the tasks without logging in
 *
 */
func unauthorized(w http.ResponseWriter, r *http.Request){
	err := templ.ExecuteTemplate(w, "unauthorized", "You must be a user to access that page.")
	if err != nil {
		fmt.Print(err.Error())
	}
}

/* Routes to user page upon successful login in
 *
 */
func userPage(w http.ResponseWriter, r *http.Request){
	isLogged, name := CheckLoginStatus(w,r)

	if isLogged {
		err := templ.ExecuteTemplate(w, "userHome", name)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		http.Redirect(w,r,"/unauthorized",http.StatusSeeOther)
	}
}

/* Checks user credentials with that of the database
 *
 */
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

//custom middleware to check if the user is logged in before executing functions, meant to wrap other functions
func isAuthenticated(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isLogged,_ := CheckLoginStatus(w, r) //first check if the user is logged in
			if isLogged { //if they are serve the function inside
				next.ServeHTTP(w,r)
			}else { //otherwise redirect to the unauthorized page
				http.Redirect(w,r,"/unauthorized",http.StatusSeeOther)
			}
		})
}

