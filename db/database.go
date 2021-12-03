package db

import (
	"database/sql"
	"log"
	"handle_api/model"
)

var database = make(map[string]interface{})

func InsertCompany(sqliteDatabase *sql.DB, name string, tel string, email string) bool {
	insertCompanySQL := `INSERT INTO company(name, tel, email) VALUES (?, ?, ?)`
	// defer sqliteDatabase.Close()
	statement, err := sqliteDatabase.Prepare(insertCompanySQL) // Prepare statement. 
                                                   // This is good to avoid SQL injections
	if err != nil {
		return false
	}
	_, err = statement.Exec(name, tel, email)
	if err != nil {
		return false
	}
	return true
}
func FindAll(sqliteDatabase *sql.DB) []interface{} {
	rows, err := sqliteDatabase.Query("SELECT COUNT(*) as count FROM  company")
	count := CheckCount(rows)
	items := make([]interface{}, 0, count)

	row, err := sqliteDatabase.Query("SELECT * FROM company ORDER BY name")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	log.Println(row)
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var tel string
		var email string
		row.Scan(&name, &tel, &email)
		com := new(model.Company)
		com.Name = name
		com.Tel = tel
		com.Email = email
		items = append(items, com)
	}
	for _, v := range items{
		log.Print("test item")
		log.Println(v)
	}
	return items
}

func FindBy(sqliteDatabase *sql.DB, key string) (interface{}, bool) {
	// com, ok := database[key]
	items := make([]interface{}, 0, 1)
	row, err := sqliteDatabase.Query("SELECT * FROM company WHERE name = '"+ key + "'")
	if err != nil {
		return make([]interface{}, 0), false
	}
	defer row.Close()
	log.Println(row)
	ok := false
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var tel string
		var email string
		row.Scan(&name, &tel, &email)
		com := new(model.Company)
		com.Name = name
		com.Tel = tel
		com.Email = email
		items = append(items, com)
		ok = true
	}
	return items, ok
}


func Update(sqliteDatabase *sql.DB, company *model.Company, root_name string)  bool{
	name :=company.Name
	tel := company.Tel
	email := company.Email
	stmt, err := sqliteDatabase.Prepare("UPDATE company set name = ?, tel = ?, email = ? where name = ?")
	if err != nil {
		return false
	}
	res, err := stmt.Exec(name, tel ,email , root_name)
	if err != nil {
		return false
	}
	affected, err := res.RowsAffected()
	if affected >=1 {
		log.Println("is affected")
		return true
	}
	return false
}
func Save(key string, item interface{}) {
	database[key] = item
}

func Remove(sqliteDatabase *sql.DB, name string)  bool {
	stmt, err := sqliteDatabase.Prepare("DELETE FROM company where name = ?")
	if err != nil {
		return false
	}
	res, err := stmt.Exec(name)
	if err != nil {
		return false
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false
	}
	if affected >=1 {
		return true
	}
	return false
}


func CheckCount(rows *sql.Rows) (count int) {
	for rows.Next() {
	   err:= rows.Scan(&count)
	   if err != nil {
		   log.Println("Error check count")
		   return 0
    	}
	}
   return count
}
