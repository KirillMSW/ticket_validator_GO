package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"log"
	"net/http"
	"os"
)

var spreadsheetId = fmt.Sprintf("%v", viper.Get("table_url"))

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

func updateStatus(ticket_id, status string) {
	readRange := "A:A"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	row_to_update := -1
	table_clients := resp.Values
	for i, a := range table_clients {
		if a[0] == ticket_id {
			row_to_update = i
		}

	}
	if row_to_update != -1 {
		row_to_update += 1
		formatRange := fmt.Sprintf("B%d", row_to_update)
		rb := &sheets.ValueRange{Values: [][]interface{}{{status}}}

		_, err = srv.Spreadsheets.Values.Update(spreadsheetId, formatRange, rb).ValueInputOption("USER_ENTERED").Do()
		if err != nil {
			fmt.Println(err)
		}
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
