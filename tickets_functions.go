package main

import (
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/skip2/go-qrcode"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

func getTickets() []Client {
	jsonFile, err := os.Open("db.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("Successfully Opened users.json")
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {

		}
	}(jsonFile)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var clients []Client
	err = json.Unmarshal(byteValue, &clients)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully unmarshled users.json")
	return clients
}

func searchClient(clients []Client, ticket_id string) int {
	for i, a := range clients {
		if a.Ticket_id == ticket_id {
			return i
		}
	}
	return -1
}

func addTicket(clients []Client, ticket_id string, people_amount int, name, surname, patronymic, phone, email string) {
	clients = append(clients, Client{
		Ticket_id:     ticket_id,
		People_amount: people_amount,
		Name:          name,
		Surname:       surname,
		Patronymic:    patronymic,
		Phone:         phone,
		Email:         email,
	})
	result, err := json.Marshal(clients) //Returns the Json encoding of u into the variable result
	if err != nil {
		fmt.Println("error", err)
	}
	err = os.WriteFile("db.json", result, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func writeTickets(clients []Client) {
	result, err := json.Marshal(clients) //Returns the Json encoding of u into the variable result
	if err != nil {
		fmt.Println("error", err)
	}
	err = os.WriteFile("db.json", result, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func generateTicketId(clients []Client) string {
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
		if searchClient(clients, ticket_id) != -1 {
			generate_again = true
		}
	}
	return ticket_id
}

func drawTicket(ticket_id, surname, name, patronymic string) {
	png, err := qrcode.New(ticket_id, qrcode.Medium)
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	code_img := png.Image(400)
	background_file, err := os.Open("qr_background.jpg")
	if err != nil {
		fmt.Println(err)
	}

	background, _, err := image.Decode(background_file)
	if err != nil {
		fmt.Println(err)
	}

	m := image.NewRGBA(background.Bounds())
	draw.Draw(m, m.Bounds(), background, image.Point{0, 0}, draw.Src)
	draw.Draw(m, m.Bounds(), code_img, image.Point{-334, -450}, draw.Src)

	dc := gg.NewContextForImage(m)
	err = dc.LoadFontFace("fonts/agencyfbcyrillic.ttf", 120)
	if err != nil {
		fmt.Println(err)
	}

	dc.SetRGB(1, 1, 1)
	dc.DrawStringAnchored(ticket_id, 534, 1350, 0.5, 1)
	dc.DrawStringWrapped(surname+"\n"+name+"\n"+patronymic, 534, 1500, 0.5, 0, 1000, 1.25, gg.AlignCenter)
	dc.Clip()
	toimg, err := os.Create("tickets/" + ticket_id + ".jpg")
	if err != nil {
		fmt.Println(err)
	}
	defer func(toimg *os.File) {
		err := toimg.Close()
		if err != nil {

		}
	}(toimg)

	err = jpeg.Encode(toimg, dc.Image(), &jpeg.Options{jpeg.DefaultQuality})
	if err != nil {
		return
	}
}
