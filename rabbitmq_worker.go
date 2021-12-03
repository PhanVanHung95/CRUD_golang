package main

import (
        "bytes"
        "github.com/streadway/amqp"
	_ "github.com/mattn/go-sqlite3"
        "log"
        "time"
        "os"
        "database/sql"
	"github.com/joho/godotenv"

)

func failOnError(err error, msg string) {
        if err != nil {
                log.Fatalf("%s: %s", msg, err)
        }
}

func insertLog(db *sql.DB, time string, info string) bool {
	insertLogSQL := `INSERT INTO log(time, info) VALUES (?, ?)`
	// defer db.Close()

	statement, err := db.Prepare(insertLogSQL) // Prepare statement. 

	if err != nil {
		return false
	}
	_, err = statement.Exec(time, info)
	if err != nil {
		return false
	}

	return true
}

func main() {
	godotenv.Load()

	var is_exist_db = false
	if _, err := os.Stat(os.Getenv("LOG_DB")); err == nil {
		is_exist_db = true
		log.Println("logs-database existed")
		// os.Remove("sqlite-database.db") 
	
	}else{
		
		file, err := os.Create(os.Getenv("LOG_DB")) // Create SQLite file
		if err != nil {
			log.Fatal(err.Error())
		}
		file.Close()
		log.Println("log-database.db created")
		// defer sqliteDatabase.Close()
	}
	logsDatabase, _ := sql.Open("sqlite3", os.Getenv("LOG_DB"))
	if is_exist_db == false{
		logsDatabase = createTable(logsDatabase) }


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
        failOnError(err, "Failed to declare a queue")

        err = ch.Qos(
                1,     // prefetch count
                0,     // prefetch size
                false, // global
        )
        failOnError(err, "Failed to set QoS")

        msgs, err := ch.Consume(
                q.Name, // queue
                "",     // consumer
                false,  // auto-ack
                false,  // exclusive
                false,  // no-local
                false,  // no-wait
                nil,    // args
        )
        failOnError(err, "Fail to register a consumer")

        forever := make(chan bool)

        go func() {
                for d := range msgs {
                        log.Printf("Received a message: %s", d.Body)
                        log.Println(bytes.NewBuffer(d.Body).String())
                        is_insert := insertLog(logsDatabase, time.Now().String(), bytes.NewBuffer(d.Body).String())
                        log.Println(is_insert)
                        dotCount := bytes.Count(d.Body, []byte("."))
                        t := time.Duration(dotCount)
                        time.Sleep(t * time.Second)
                        d.Ack(false)
                }
        }()

        log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
        <-forever
        }

func createTable(db *sql.DB) *sql.DB {
        createLogTableSQL := `CREATE TABLE log (
                "time" TEXT NOT NULL PRIMARY KEY,		
                "info" TEXT
                );` // SQL Statement for Create Table

        log.Println("Create log table...")
        statement, err := db.Prepare(createLogTableSQL) // Prepare SQL Statement
        if err != nil {
                log.Fatal(err.Error())
        }
        statement.Exec() // Execute SQL Statements
        return db
}