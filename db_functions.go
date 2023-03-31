package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
	"math/rand"
	"time"
)

func connectToDatabase() {
	cfg := mysql.Config{
		User:   fmt.Sprintf("%v", viper.Get("db_user")),
		Passwd: fmt.Sprintf("%v", viper.Get("db_password")),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "new_schema",
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
}

func dbGenerateTicketId() string {
	rand.Seed(time.Now().UnixNano())
	generate_again := true
	var ticket_id string
	for generate_again {
		generate_again = false
		b := make([]byte, 5)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}
		ticket_id = string(b)
		result := db.QueryRow("SELECT ticket_id FROM clients WHERE ticket_id=?", ticket_id)
		err := result.Scan()
		if err == nil {
			generate_again = true
		}

	}
	return ticket_id
}
func dbAddTicket(ticket_id string, people_amount int, name, surname, patronymic, phone, email string) {
	_, err := db.Exec(
		"INSERT INTO clients VALUES (?, ?, ?, ?, ?, ?, ?)",
		ticket_id,
		people_amount,
		surname,
		name,
		patronymic,
		phone,
		email,
	)
	if err != nil {

	}
}

func dbGetTicket(ticket_id string) (Client, error) {
	result := db.QueryRow("SELECT * FROM clients WHERE ticket_id=?", ticket_id)
	var client Client
	err := result.Scan(
		&client.Ticket_id,
		&client.People_amount,
		&client.Surname,
		&client.Name,
		&client.Patronymic,
		&client.Phone,
		&client.Email,
	)
	if err != nil {
		fmt.Println(err)
		return Client{}, errors.New("No rows")
	}
	return client, nil
}

func dbUpdatePeopleAmount(new_value int, ticket_id string) error {
	_, err := db.Exec("UPDATE clients SET people_amount=? WHERE ticket_id=?", new_value, ticket_id)
	if err != nil {
		fmt.Println(err)
		return errors.New("No rows")
	}
	return nil
}
