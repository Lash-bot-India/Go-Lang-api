package main

import (
	"database/sql"
	"fmt"

	"crypto/rand"
	"math/big"
	"strconv"

	"strings"

	"time"

	_ "github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
)

const (
	keyList string = "abcdefghijklmnopqrstuvwxyzABCDEFHFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type LicenceKey struct {
	Lkey string `json:"lkey"`
}

type Licence struct {
	LicenceKey    string `json:"licencekey"`
	ExpiryDate    string `json:"expirydate"`
	ClientName    string `json:"clientname"`
	LoginStatus   string `json:"loginstatus"`
	LicenceStatus string `json:"licencestatus"`
}

type ClientInfo struct {
	ClientID   int    `json:"clientid"`
	ExpiryDate string `json:expirydate`
}

func generatelicencekey() string {
	var lkey string
	var tmpchar string
	size := "16"
	strLen, _ := strconv.Atoi(size)
	for key := 1; key <= strLen; key++ {
		res, _ := rand.Int(rand.Reader, big.NewInt(64))
		keyGen := keyList[res.Int64()]
		stringGen := fmt.Sprintf("%c", keyGen)
		tmpchar = tmpchar + string([]byte(stringGen))
	}
	lkey = strings.ToUpper(tmpchar)
	return lkey

}
func licenceagenerate(client *Client, data interface{}) {
	var clientinfo ClientInfo
	mapstructure.Decode(data, &clientinfo)
	var lickey = generatelicencekey()
	go func() {
		_, queryerr := client.session.Exec("INSERT INTO licencemstr(licencekey, clientid, loginstatus, expirydate, licencestatus) VALUES($1,$2,$3,$4, $5);", lickey, clientinfo.ClientID, "inactive", clientinfo.ExpiryDate, "inactive")
		if queryerr != nil {
			fmt.Println(queryerr)
			client.send <- Message{"error", queryerr}
		} else {
			client.send <- Message{"success", lickey}
		}
	}()
}

func licenceactivate(client *Client, data interface{}) {
	var FirstNAME string
	var LastNAME string
	var ClientID int
	var lkey LicenceKey
	var licence Licence
	var message Message
	mapstructure.Decode(data, &lkey)
	go func() {
		row := client.session.QueryRow("select licencekey, clientid, loginstatus, licencestatus, expirydate from licencemstr where licencekey=$1;", lkey.Lkey)
		switch err := row.Scan(&licence.LicenceKey, &ClientID, &licence.LoginStatus, &licence.LicenceStatus, &licence.ExpiryDate); err {
		case sql.ErrNoRows:
			message.Name = "error"
			message.Data = "Invalid Licence Key"
			client.send <- message
			//fmt.Println("No rows were returned!")
		case nil:
			if licence.LicenceStatus == "active" {
				message.Name = "error"
				message.Data = "This licence is already in use"
				client.send <- message
			} else {
				sqlStatement := `UPDATE licencemstr SET licencestatus = 'active', loginstatus='active' WHERE licencekey = $1;`
				_, err = client.session.Exec(sqlStatement, lkey.Lkey)
				row2 := client.session.QueryRow("select fname, lname from clientmstr where clientid=$1;", ClientID)
				row2.Scan(&FirstNAME, &LastNAME)
				licence.ClientName = FirstNAME + " " + LastNAME
				licence.LoginStatus = "active"
				licence.LicenceStatus = "inactive"
				if err != nil {
					message.Name = "error"
					message.Data = licence
					client.send <- message
				} else {
					message.Name = "success"
					message.Data = licence
					client.send <- message
				}
			}
		default:
			panic(err)
		}
	}()
}

/*
func adduser(client *Client, data interface{}) {
	var user User
	var message Message
	mapstructure.Decode(data, &user)
	var lastInsertID int
	fmt.Printf("%#v\n", user)
	go func() {
		//sqlStatement := `INSERT INTO user_mstr (username, password, firstname, lastname) VALUES ('admin', 'admin', 'admin', 'admin')`
		//_, queryerr := client.session.Exec(sqlStatement)

		queryerr := client.session.QueryRow("INSERT INTO user_mstr(username, password, firstname, lastname) VALUES($1,$2,$3,$4) RETURNING userid;", user.Username, user.Password, user.Firstname, user.Lastname).Scan(&lastInsertID)
		fmt.Println(lastInsertID)
		if queryerr != nil {
			client.send <- Message{"error", queryerr}
		}

	}()
	user.UserID = lastInsertID
	user.Firstname = user.Firstname
	user.Lastname = user.Lastname
	fmt.Printf("%#v\n", user)
	message.Name = "user add"
	message.Data = user
	client.send <- message
}
*/
func testsocket(client *Client, data interface{}) {
	var lkey LicenceKey
	var message Message
	mapstructure.Decode(data, &lkey)
	message.Name = "returning"
	message.Data = lkey
	client.send <- message

}

func licencevalidation(client *Client, data interface{}) {
	var lkey LicenceKey
	var licencestatus Licence
	var message Message

	mapstructure.Decode(data, &lkey)
	go func() {
		row := client.session.QueryRow("select loginstatus, licencestatus, expirydate from licencemstr where licencekey=$1;", lkey.Lkey)
		switch err := row.Scan(&licencestatus.LoginStatus, &licencestatus.LicenceStatus, &licencestatus.ExpiryDate); err {
		case sql.ErrNoRows:
			message.Name = "error"
			message.Data = "Invalid Licence Key"
			client.send <- message
		case nil:
			if licencestatus.LicenceStatus == "inactive" {
				t, _ := time.Parse(time.RFC3339, licencestatus.ExpiryDate)
				fmt.Println(t)
				currentTime := time.Now()
				datediff := currentTime.Sub(t)
				if datediff > 0 {
					message.Name = "error"
					message.Data = "Licence Key Has been expired.\n Please Renew your Licence."
					client.send <- message
				} else {
					message.Name = "error"
					message.Data = "Please Reactivate your Licence"
					client.send <- message
				}

			} else {
				if licencestatus.LoginStatus == "active" {
					message.Name = "error"
					message.Data = "Sorry another instance is running"
					client.send <- message
				} else {
					message.Name = "success"
					message.Data = "Validated in Successfully"
					client.send <- message
				}
			}
		default:
			panic(err)
		}
	}()
}

func login(client *Client, data interface{}) {
	var creds Credential
	var user User
	var message Message
	mapstructure.Decode(data, &creds)
	go func() {
		row := client.session.QueryRow("select firstname, lastname from user_mstr where username=$1 and password=$2;", creds.Username, creds.Password)
		switch err := row.Scan(&user.Firstname, &user.Lastname); err {
		case sql.ErrNoRows:
			message.Name = "error"
			message.Data = "No rows were returned!"
			client.send <- message
			//fmt.Println("No rows were returned!")
		case nil:
			message.Name = "success"
			message.Data = "No rows were returned!"
			client.send <- message
			fmt.Println(user.Firstname, user.Lastname)
		default:
			panic(err)
		}
	}()

	//creds.Username = lastInsertId
	//user.Firstname = user.Firstname
	//user.Lastname = user.Lastname
	//fmt.Printf("%#v\n", user)

}
