package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"database/sql"

	"github.com/beeker1121/goque"
	_ "github.com/mattn/go-sqlite3"
)

var Q, err = goque.OpenQueue("data_dir")
var db, err1 = sql.Open("sqlite3", "database.db")

func queryParamDisplayHandler(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "id: "+req.FormValue("id"))
	initialstatus := "Open"
	sqlDatabase(req.FormValue("id"), initialstatus)
	queueDataHandler(req.FormValue("id"))
	queryResponse(res, req.FormValue("id"), req)

}

func queryResponse(res http.ResponseWriter, id string, req *http.Request) {
	flag := 0
	check := func(res http.ResponseWriter, id string, stop chan bool) int {
		var status string

		err = db.QueryRow("select status from Done where id = ?", id).Scan(&status)

		if status == "Done" && flag == 0 {

			io.WriteString(res, " Status: Done")
			flag = 1

			stop <- true
			return 1
		}
		return 0
	}
	stop := make(chan bool)
	stop = schedule(id, res, check, 1*time.Millisecond, stop)
	msg := <-stop
	fmt.Println(msg)
}
func schedule(id string, res http.ResponseWriter, what func(res http.ResponseWriter, id string, stop chan bool) int, delay time.Duration, stop chan bool) chan bool {

	var n int
	n = 0
	go func(res http.ResponseWriter, id string, stop chan bool) {
		for {
			what(res, id, stop)

			select {
			case <-time.After(delay):
				n = n + 1

				if n == 4999 {
					io.WriteString(res, " Status: Processing")
					stop <- true
					return
				}
			case <-stop:
				return
			}
		}
	}(res, id, stop)

	return stop
}
func queryParamStatusHandler(res http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")

	io.WriteString(res, "status: "+lookupDatabase(id))
}

func lookupDatabase(id string) string {

	var status string
	err = db.QueryRow("select status from Done where id = ?", id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("SORRY WRONG ID")
			// there were no rows, but otherwise no error occurred
		} else {
			log.Fatal(err)
		}
	}
	return status
}
func queueDataHandler(id string) {
	//Q, err := goque.OpenQueue("data_dir")
	Q.Enqueue([]byte(id))

	//defer Q.Close()
}

func statusHandler() {
	done := make(chan string)
	//Q, err := goque.OpenQueue("data_dir")

	for true {
	

		item, err := Q.Peek()
		//fmt.Println(err)
		if err != nil {
			//fmt.Println("In error")
		} else {
			stmt, err := db.Prepare("UPDATE Done SET status='Processing' WHERE id=?")
			if err != nil {
				log.Fatal(err)
			}
			_, err = stmt.Exec(item.ToString())
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Println(item.ToString())
			go apiWorker(done, item.ToString())
			msg := <-done
			item, err = Q.Dequeue()
			fmt.Println(msg)

		}
		//defer Q.Close()
	}
}

func apiWorker(done chan string, id string) {
	fmt.Println("Working with id ", id)

	// call your api here

	time.Sleep(time.Duration(rand.Intn(8000)) * time.Millisecond) // this is to just stop for random time remove the line

	stmt, err := db.Prepare("UPDATE Done SET status='Done' WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(id)
	if err != nil {
		log.Fatal(err)
	}

	done <- "Done" // We send a message on the channel
}
func sqlDatabase(id string, status string) {

	if err1 != nil {
		panic(err.Error())
	}
	stmt, err := db.Prepare("INSERT INTO Done(id,status) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(id, status)
	if err != nil {
		log.Fatal(err)
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)

	if err != nil {
		panic(err.Error())
	}

}

func main() {

	http.HandleFunc("/validatejob", func(res http.ResponseWriter, req *http.Request) {
		queryParamDisplayHandler(res, req)
	})
	http.HandleFunc("/getjobstatus", func(res http.ResponseWriter, req *http.Request) {
		queryParamStatusHandler(res, req)
	})
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS Done (id VARCHAR(255),status VARCHAR(255))")
	statement.Exec()
	println("Enter this in your browser:  http://localhost:8080/validatejob?id=6")
	println("Enter this in your browser:  http://localhost:8080/getjobstatus?id=6")

	go statusHandler()

	http.ListenAndServe(":8080", nil)
}
