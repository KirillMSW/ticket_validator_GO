package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
	"strconv"
)

// album represents data about a record album.
type Client struct {
	Ticket_id     string `json:"ticket_id"`
	People_amount int    `json:"people_amount"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Patronymic    string `json:"patronymic"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
}

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var srv *sheets.Service

var db *sql.DB

func main() {
	onStartup()
	client, err := dbGetTicket("kkjk")
	if err != nil {
		fmt.Println(err)
	}
	client.Name = "s"
	fmt.Println("Begining app")
	router := gin.Default()
	router.POST("/api/generate", generate)
	router.GET("/api/validate", validate)
	router.POST("/api/checkin", checkin)
	router.POST("/api/void", void_ticket)
	router.Run("localhost:8080")
}

func onStartup() {
	viper.SetConfigFile("secrets.json")
	err := viper.ReadInConfig()
	if err != nil {
		return
	}
	connectToDatabase()
	ctx := context.Background()
	config := setConfig()
	client := getSheetsClient(config)
	srv = getService(ctx, client)
}

//func checkFiles() error {
//	jsonFile, err := os.Open("db.json")
//	if err != nil {
//		fmt.Println("Failed to open db.json")
//		fmt.Println(err)
//		return err
//	}
//	fmt.Println("Successfully opened db.json")
//	defer func(jsonFile *os.File) {
//		err := jsonFile.Close()
//		if err != nil {
//
//		}
//	}(jsonFile)
//
//}

func generate(c *gin.Context) {
	people_amount := 1
	name := c.PostForm("name")
	surname := c.PostForm("surname")
	patronymic := c.PostForm("patronymic")
	phone := c.PostForm("phone")
	email := c.PostForm("email")

	clients := getTickets()
	ticket_id := generateTicketId(clients)

	addTicket(clients, ticket_id, people_amount, name, surname, patronymic, phone, email)
	go pushToTables(ticket_id, surname, name, patronymic, phone, email)
	drawTicket(ticket_id, surname, name, patronymic)

	dbAddTicket(ticket_id, people_amount, name, surname, patronymic, phone, email)
	c.Header("Access-Control-Allow-Origin", "*")
	c.File("tickets/" + ticket_id + ".jpg")

}

func checkin(c *gin.Context) {
	ticket_id := c.PostForm("ticket_id")
	people_to_pass, err := strconv.Atoi(c.PostForm("people_to_pass"))
	if err != nil {
		fmt.Println(err)
	}
	clients := getTickets()
	client_num_id := searchClient(clients, ticket_id)

	c.Header("Access-Control-Allow-Origin", "*")
	if client_num_id != -1 {
		clients[client_num_id].People_amount -= people_to_pass
		writeTickets(clients)
		go updateStatus(ticket_id, "ИЗРАСХОДОВАН")
		c.String(200, "OK")
	} else {
		c.String(200, "INVALID")
	}

}

func validate(c *gin.Context) {
	ticket_id := c.Query("key")
	client, err := dbGetTicket(ticket_id)
	c.Header("Access-Control-Allow-Origin", "*")
	if err != nil {
		c.JSON(200, "{}")
	} else {
		c.JSON(200, client)
	}
}

func void_ticket(c *gin.Context) {
	ticket_id := c.PostForm("ticket_id")
	clients := getTickets()
	client_num_id := searchClient(clients, ticket_id)

	c.Header("Access-Control-Allow-Origin", "*")
	if client_num_id != -1 {
		clients[client_num_id].People_amount = -1
		writeTickets(clients)
		go updateStatus(ticket_id, "АННУЛИРОВАН")
		c.String(200, "OK")
	} else {
		c.String(200, "INVALID")
	}
}
