package main

import (
	"github.com/gorilla/mux"
	"handle_api/handler"
	"log"
	"net/http"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"os"
	"github.com/streadway/amqp"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	var is_exist_db = false
	if _, err := os.Stat(os.Getenv("CRUD_DB")); err == nil {
		is_exist_db = true
		log.Println("Sqlite-database existed")
		// os.Remove("sqlite-database.db") 
	
	}else{
		
		file, err := os.Create(os.Getenv("CRUD_DB")) // Create SQLite file
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
		log.Println("sqlite-database created")
		// defer sqliteDatabase.Close()
	}
	sqliteDatabase, _ := sql.Open("sqlite3", os.Getenv("CRUD_DB"))
	if is_exist_db == false{
		sqliteDatabase = createTable(sqliteDatabase) }
		conn, err := amqp.Dial(os.Getenv("LINK_RABITMQ"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
			"task_queue", // name
			true,         // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
	)

	router := mux.NewRouter().StrictSlash(true)
	sub := router.PathPrefix("/api/v1").Subrouter()
	sub.Methods("GET").Path("/companies").HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		handler.GetCompanies(w,r, sqliteDatabase, ch, q)
	})
	sub.Methods("POST").Path("/companies").HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		handler.SaveCompany(w, r, sqliteDatabase, ch, q)})

	sub.Methods("GET").Path("/companies/{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			handler.GetCompany(w, r, sqliteDatabase, ch, q)})
	sub.Methods("PUT").Path("/companies/{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		handler.UpdateCompany(w, r, sqliteDatabase, ch, q)})

	sub.Methods("DELETE").Path("/companies/{name}").HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		handler.DeleteCompany(w, r, sqliteDatabase, ch, q)})
	log.Fatal(http.ListenAndServe(":" +  os.Getenv("PORT"), router))
}

func createTable(db *sql.DB) *sql.DB {
	createStudentTableSQL := `CREATE TABLE company (
		"name" TEXT NOT NULL PRIMARY KEY,		
		"tel" TEXT,
		"email" TEXT
	  );` // SQL Statement for Create Table

	log.Println("Create Company table...")
	statement, err := db.Prepare(createStudentTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	return db
}

func failOnError(err error, msg string) {
	if err != nil {
			log.Fatalf("%s: %s", msg, err)
	}
}