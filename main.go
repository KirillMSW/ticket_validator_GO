package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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

var srv *sheets.Service

func main() {
	onStartup()
	fmt.Println("Begining app")
	router := gin.Default()
	router.POST("/api/generate", generate)
	router.GET("/api/validate", validate)
	router.POST("/api/checkin", checkin)
	router.Run("localhost:8080")
}

func onStartup() {
	ctx := context.Background()
	config := setConfig()
	client := getSheetsClient(config)
	srv = getService(ctx, client)
}

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
		fmt.Printf("%d", i)
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

func pushToTables(ticket_id, surname, name, patronymic, phone, email string) {
	readRange := "A:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	new_row_id := len(resp.Values) + 1
	formatRange := fmt.Sprintf("A%d:G%d", new_row_id, new_row_id)
	rb := &sheets.ValueRange{Values: [][]interface{}{{ticket_id, "АКТИВЕН", surname, name, patronymic, phone, email}}}

	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, formatRange, rb).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		fmt.Println(err)
	}

}

func drawTicket(ticket_id, surname, name, patronymic string) {
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

	dc.SetRGB(1, 1, 1)
	dc.DrawStringAnchored(ticket_id, 534, 1350, 0.5, 1)
	dc.DrawStringWrapped(name+"\n"+surname+"\n"+patronymic, 534, 1500, 0.5, 0, 1000, 1.25, gg.AlignCenter)
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
func getService(ctx context.Context, client *http.Client) *sheets.Service {
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return srv
}
func setConfig() *oauth2.Config {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}
func getSheetsClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return
	}
}

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
		c.String(200, "OK")
	} else {
		c.String(200, "INVALID")
	}

}

func validate(c *gin.Context) {
	ticket_id := c.Query("key")

	clients := getTickets()
	client_num_id := searchClient(clients, ticket_id)

	c.Header("Access-Control-Allow-Origin", "*")
	if client_num_id != -1 {
		client := clients[client_num_id]
		c.JSON(200, client)
	} else {
		c.JSON(200, "{}")
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
		c.String(200, "OK")
	} else {
		c.String(200, "INVALID")
	}
}
