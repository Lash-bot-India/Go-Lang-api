package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	//host := "licensemgrdb.postgres.database.azure.com"
	host := "localhost"
	port := 5432
	//user := "lashbot@licensemgrdb"
	user := "postgres"
	//password := "createthebot@2020"
	password := "postgres"
	//dbname := "lashbotdb"
	dbname := "lashbot"
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
	router.Handle("test", testsocket)
	router.Handle("validate licence", licencevalidation)
	router.Handle("activate licence", licenceactivate)
	router.Handle("generate licence", licenceagenerate)

	http.Handle("/", router)
	//http.HandleFunc("/", handler)
	//http.ListenAndServeTLS()
	//http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
	http.ListenAndServe(":4000", nil)

}
