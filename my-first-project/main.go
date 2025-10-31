package main

import (
	"fmt"
	"net/http"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func main(){

	connectDB()

	

	//map the handler function for root 
	http.HandleFunc("/", homeHandler)
    	http.HandleFunc("/health", healthcheckHandler)
	
	fmt.Println("Server initiated :", "http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}


func connectDB(){
	var err error
	dsn:= "host=localhost user=myuser password=mypassword dbname=db port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	if err != nil {
		log.Fatal("DB connection failed: ", err)
	}
	
	fmt.Println("DB connection successful!")
}

func homeHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Server working!\n")
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request){
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintln(w, "DB connection falied\n")	
		return
	}
	
	err = sqlDB.Ping()
    	if err != nil {
        	fmt.Fprintf(w, "❌ DB Ping 실패\n")
        	return
    	}
	fmt.Fprintf(w, "Server healthy\n DB connection healthy")

}



