package main

import (
	"net/http"
	"fmt"
	"log"
	"encoding/json"
)

func GetTasks(w http.ResponseWriter, r *http.Request){

	isLogged, _ := CheckLoginStatus(w,r)

	if !isLogged{
		http.Redirect(w,r,"/",http.StatusUnauthorized)
	}

	type Task struct {
		TaskId int
		TaskName string
		TaskDesc string
	}

	var t_id int
	var t_name, t_info string

	task_list := make([]Task,0)

	rows, _ := db.Query("SELECT * FROM tasks")

	for rows.Next(){
		if rows.Scan(&t_id,&t_name,&t_info) ; err != nil{
			log.Fatalln(err.Error())
		}

		task := Task{TaskId:t_id, TaskName:t_name, TaskDesc:t_info}

		task_list = append(task_list, task)
	}
	defer rows.Close()


	data , err := json.Marshal(task_list)
	if err != nil {
		fmt.Println(err.Error())
	}

	//write those bytes to the response to use for on client side
	w.Write(data)
	fmt.Println(string(data))
}

func AddTask(w http.ResponseWriter, r *http.Request){
	fmt.Println("Adding task")

	task_name := r.FormValue("task")
	task_info := r.FormValue("info")

	fmt.Println("Task name",task_name)
	fmt.Println("Task info",task_info)


	_ ,err := db.Exec("INSERT INTO tasks ( task_name , task_desc ) values ( $1 , $2 )",task_name,task_info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
