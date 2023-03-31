package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/api/sheets/v4"
	"os"
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

var spreadsheetId string

func main() {
	err := onStartup()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Begining app")
	router := gin.Default()
	router.POST("/api/generate", generate)
	router.GET("/api/validate", validate)
	router.POST("/api/checkin", checkin)
	router.POST("/api/void", void_ticket)
	router.Run("localhost:8080")
}

func onStartup() error {
	err := checkFiles()
	if err != nil {
		fmt.Println(err)
		return errors.New("No required files")
	}
	viper.SetConfigFile("secrets.json")
	err = viper.ReadInConfig()
	if err != nil {
		return errors.New("Failed to read config")
	}
	connectToDatabase()
	spreadsheetId = fmt.Sprintf("%v", viper.Get("table_url"))
	ctx := context.Background()
	config := setConfig()
	client := getSheetsClient(config)
	srv = getService(ctx, client)
	return nil
}

//TODO Check for: font file

func checkFiles() error {
	if _, err := os.Stat("db.json"); !os.IsNotExist(err) {
		fmt.Println("db.json exists")
	} else {
		fmt.Println("db.json does not exist. Creating empty file")
		result, err := json.Marshal("{}")
		if err != nil {
			fmt.Println("error", err)
		}
		err = os.WriteFile("db.json", result, 0644)
		if err != nil {
			fmt.Println(err)
		}
	}

	if _, err := os.Stat("qr_background.jpg"); !os.IsNotExist(err) {
		fmt.Println("qr_background.jpg exists")
	} else {
		return errors.New("qr_background.jpg does not exist")
	}

	if _, err := os.Stat("secrets.json"); !os.IsNotExist(err) {
		fmt.Println("secrets.json exists")
	} else {
		return errors.New("secrets.json does not exist")
	}

	if _, err := os.Stat("credentials.json"); !os.IsNotExist(err) {
		fmt.Println("credentials.json exists")
	} else {
		return errors.New("credentials.json does not exist")
	}

	dir_info, err := os.Stat("tickets")
	if !os.IsNotExist(err) && dir_info.IsDir() {
		fmt.Println("tickets directory exists")
	} else {

		err := os.Mkdir("tickets", 0755)
		if err != nil {
			return err
		}
		fmt.Println("tickets directory does not exist. Creating")
	}

	dir_info, err = os.Stat("fonts")
	if !os.IsNotExist(err) && dir_info.IsDir() {
		fmt.Println("fonts directory exists")
	} else {
		return errors.New("fonts directory does not exist")
	}

	return nil
}

func generate(c *gin.Context) {
	people_amount := 1
	name := c.PostForm("name")
	surname := c.PostForm("surname")
	patronymic := c.PostForm("patronymic")
	phone := c.PostForm("phone")
	email := c.PostForm("email")

	//clients := getTickets()
	ticket_id := dbGenerateTicketId()

	dbAddTicket(ticket_id, people_amount, name, surname, patronymic, phone, email)

	//addTicket(clients, ticket_id, people_amount, name, surname, patronymic, phone, email)
	go pushToTables(ticket_id, surname, name, patronymic, phone, email)
	drawTicket(ticket_id, surname, name, patronymic)

	dbAddTicket(ticket_id, people_amount, name, surname, patronymic, phone, email)
	c.Header("Access-Control-Allow-Origin", "*")
	c.File("tickets/" + ticket_id + ".jpg")

}

// TODO: optimize (unite search and changing)
func checkin(c *gin.Context) {
	ticket_id := c.PostForm("ticket_id")
	people_to_pass, err := strconv.Atoi(c.PostForm("people_to_pass"))
	if err != nil {
		fmt.Println(err)
	}

	//clients := getTickets()
	//client_num_id := searchClient(clients, ticket_id)

	client, err := dbGetTicket(ticket_id)
	c.Header("Access-Control-Allow-Origin", "*")
	if err == nil {
		//clients[client_num_id].People_amount -= people_to_pass
		//writeTickets(clients)
		err = dbUpdatePeopleAmount(client.People_amount-people_to_pass, ticket_id)
		if err != nil {
			c.String(200, "INVALID")
		}
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
		if err.Error() == "No rows" {
			fmt.Println("aboba")
		}
		c.JSON(200, "{}")
		fmt.Println(err)
	} else {
		c.JSON(200, client)
	}
}

func void_ticket(c *gin.Context) {
	ticket_id := c.PostForm("ticket_id")
	//clients := getTickets()
	//client_num_id := searchClient(clients, ticket_id)

	c.Header("Access-Control-Allow-Origin", "*")
	err := dbUpdatePeopleAmount(-1, ticket_id)
	if err != nil {
		c.String(200, "INVALID")
		return
	}
	go updateStatus(ticket_id, "АННУЛИРОВАН")
	c.String(200, "OK")
}
