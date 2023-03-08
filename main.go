package main

import (
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
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

func main() {
	router := gin.Default()
	router.POST("/api/generate", generate)
	router.GET("/api/validate", validate)
	router.Run("localhost:8080")
}

////// getAlbums responds with the list of all albums as JSON.
////func getAlbums(c *gin.Context) {
////	c.IndentedJSON(http.StatusOK, albums)
////}
////
////// postAlbums adds an album from JSON received in the request body.
////func postAlbums(c *gin.Context) {
////	var newAlbum album
////
////	// Call BindJSON to bind the received JSON to
////	// newAlbum.
////	if err := c.BindJSON(&newAlbum); err != nil {
////		return
////	}
////
////	// Add the new album to the slice.
////	albums = append(albums, newAlbum)
////	c.IndentedJSON(http.StatusCreated, newAlbum)
////}
//
//// getAlbumByID locates the album whose ID value matches the id
//// parameter sent by the client, then returns that album as a response.
//func getAlbumByID(c *gin.Context) {
//	id := c.Param("id")
//
//	// Loop through the list of albums, looking for
//	// an album whose ID value matches the parameter.
//	for _, a := range albums {
//		if a.ID == id {
//			c.IndentedJSON(http.StatusOK, a)
//			return
//		}
//	}
//	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
//}

func generate(c *gin.Context) {
	people_amount := 1
	name := c.PostForm("name")
	surname := c.PostForm("surname")
	patronymic := c.PostForm("patronymic")
	phone := c.PostForm("phone")
	email := c.PostForm("email")

	fmt.Println(people_amount, name, surname, patronymic, phone, email)
	//var png []byte

	jsonFile, err := os.Open("db.json")
	if err != nil {
		fmt.Println(err)
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
		for _, a := range clients {
			if a.Ticket_id == ticket_id {
				generate_again = true
			}
		}
	}

	clients = append(clients, Client{
		Ticket_id:     ticket_id,
		People_amount: 1,
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
	png, err := qrcode.New(ticket_id, qrcode.Medium)
	if err != nil {
		fmt.Println(err)
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

	//S := 1000.0
	dc.SetRGB(1, 1, 1)
	dc.DrawStringAnchored(ticket_id, 534, 1350, 0.5, 1)

	dc.DrawStringWrapped(name+"\n"+surname+"\n"+patronymic, 534, 1500, 0.5, 0, 1000, 1.25, gg.AlignCenter)
	dc.Clip()
	//err = dc.SaveJPG("asdasd.png")
	//if err != nil {
	//	fmt.Println(err)
	//}
	toimg, _ := os.Create("new.jpg")
	defer func(toimg *os.File) {
		err := toimg.Close()
		if err != nil {

		}
	}(toimg)

	err = jpeg.Encode(toimg, dc.Image(), &jpeg.Options{jpeg.DefaultQuality})
	if err != nil {
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.File("new.jpg")

}
func validate(c *gin.Context) {
	id := c.Query("key")
	fmt.Println(id)
	jsonFile, err := os.Open("db.json")
	if err != nil {
		fmt.Println(err)
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

	for _, a := range clients {
		if a.Ticket_id == id {
			c.Header("Access-Control-Allow-Origin", "*")
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusNotFound, gin.H{})
}
