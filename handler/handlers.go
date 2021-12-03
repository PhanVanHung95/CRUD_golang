package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"handle_api/db"
	"handle_api/model"
	"io/ioutil"
	"net/http"
	"database/sql"
	"log"
	"github.com/streadway/amqp"

)


func PublishChannel(channel *amqp.Channel, queue amqp.Queue, content string) bool{
	err := channel.Publish(
		"",           // exchange
		queue.Name,       // routing key
		false,        // mandatory
		false,
		amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         []byte(content),
	})
	if err != nil {
		log.Println("Failed to publish a message")
		return false
	}
	return true
}

func GetCompanies(w http.ResponseWriter, _ *http.Request, sqliteDatabase *sql.DB, channel *amqp.Channel, queue amqp.Queue) {
	companies := db.FindAll(sqliteDatabase)
	PublishChannel(channel, queue, "Start to get companies")

	bytes, err := json.Marshal(companies)
	if err != nil {
		PublishChannel(channel, queue, "Error happened in JSON marshal.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	writeJsonResponse(w, bytes)
	PublishChannel(channel, queue, "Success to get companies")
}

func GetCompany(w http.ResponseWriter, r *http.Request, sqliteDatabase *sql.DB,  channel *amqp.Channel, queue amqp.Queue) {
	vars := mux.Vars(r)
	name := vars["name"]
	PublishChannel(channel, queue, "Start get " + name + " company")
	com, ok := db.FindBy(sqliteDatabase,name)
	log.Println(com, ok)
	if !ok {
		http.NotFound(w, r)
		return
	}
	bytes, err := json.Marshal(com)
	if err != nil {
		PublishChannel(channel, queue, "Error happened in JSON marshal.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	writeJsonResponse(w, bytes)
	PublishChannel(channel, queue, "Success to get info companie " + name )
}

func SaveCompany(w http.ResponseWriter, r *http.Request, sqliteDatabase *sql.DB,  channel *amqp.Channel, queue amqp.Queue) {
	PublishChannel(channel, queue, "Start to create new company")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	com := new(model.Company)
	err = json.Unmarshal(body, com)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Println(com)
	content, err:= json.Marshal(com)
	PublishChannel(channel, queue, "Info comany: " + string(content))
	if err != nil {
		PublishChannel(channel, queue, "Error happened in JSON marshal.")
	}
	is_insert := db.InsertCompany(sqliteDatabase, com.Name, com.Tel, com.Email)
	log.Println(is_insert) 
	w.Header().Set("Location", r.URL.Path+"/"+com.Name)
	result := make(map[string]string)
	
	if is_insert == true{
		result["status"] = "Success"
		PublishChannel(channel, queue, "Success create company: " + com.Name)
	}else{
		result["status"] = "Failure"
		PublishChannel(channel, queue, "Failure create company: " + com.Name)

	}
	jsonResp, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		PublishChannel(channel, queue, "Error happened in JSON marshal.")
	}
	writeJsonResponse(w, jsonResp)
}

func UpdateCompany(w http.ResponseWriter, r *http.Request, sqliteDatabase *sql.DB,  channel *amqp.Channel, queue amqp.Queue) {
	vars := mux.Vars(r)
	name := vars["name"]
	PublishChannel(channel, queue, "Start to update company name:" + name)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	com := new(model.Company)
	err = json.Unmarshal(body, com)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	is_update := db.Update(sqliteDatabase, com, name)
	result := make(map[string]string)
	
	if is_update == true{
		result["status"] = "Success"
		PublishChannel(channel, queue, "Company: " + com.Name + " update success")
	}else{
		result["status"] = "Failure"
		PublishChannel(channel, queue, "Company: " + com.Name + " update fail")
	}
	jsonResp, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		PublishChannel(channel, queue, "Error happened in JSON marshal.")
	}
	writeJsonResponse(w, jsonResp)

}

func DeleteCompany(w http.ResponseWriter, r *http.Request, sqliteDatabase *sql.DB,  channel *amqp.Channel, queue amqp.Queue) {
	vars := mux.Vars(r)
	name := vars["name"]
	PublishChannel(channel, queue, "Start to remove company " + name)
	is_remove := db.Remove(sqliteDatabase, name)
	result := make(map[string]string)
	
	if is_remove == true{
		result["status"] = "Success"
		PublishChannel(channel, queue, "Company: " + name + " remove success")
	}else{
		result["status"] = "Failure"
		PublishChannel(channel, queue, "Company: " + name + " remove fail")
	}
	jsonResp, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		PublishChannel(channel, queue, "Error happened in JSON marshal.")
	}
	writeJsonResponse(w, jsonResp)
}

func writeJsonResponse(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bytes)
}
