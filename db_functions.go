package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
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
