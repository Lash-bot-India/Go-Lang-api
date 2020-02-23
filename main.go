package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	host := "licensemgrdb.postgres.database.azure.com"
	port := 5432
	user := "lashbot@licensemgrdb"
	password := "createthebot@2020"
	dbname := "lashbotdb"
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	dbconn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connecttion Established")
	}

	router := NewRouter(dbconn)
	//router.Handle("user add", adduser)
	router.Handle("login", login)
	//router.Handle("lcvalidation", licencevalidation)
	router.Handle("activate licence", licenceactivate)

	http.Handle("/", router)
	//http.HandleFunc("/", handler)
	//http.ListenAndServeTLS()
	http.ListenAndServe(":4000", nil)
}
