package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/gin-gonic/gin"
)

type userInput struct {
	Name          string `json:"name"`
	Age           int    `json:"age"`
	CommuteMethod string `json:"commute_method"`
	College       string `json:"college"`
	Hobbies       string `json:"hobbies"`
}

var userInputs = []userInput{
	{Name: "Jack", Age: 21, CommuteMethod: "Bike", College: "Boston University", Hobbies: "Golf"},
	{Name: "David", Age: 21, CommuteMethod: "Bike", College: "Boston University", Hobbies: "Golf"},
	{Name: "Austin", Age: 21, CommuteMethod: "Bike", College: "Boston University", Hobbies: "Golf"},
}

func getUsers(context *gin.Context) {
	context.IndentedJSON(http.StatusOK, userInputs)
}

func addUser(context *gin.Context) {
	var newUser userInput

	if err := context.BindJSON(&newUser); err != nil {
		return
	}

	userInputs = append(userInputs, newUser)
	if err := addUsertoGoogleSheets(newUser); err != nil {
		context.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to store data in Google Sheets"})
		return
	}

	context.IndentedJSON(http.StatusCreated, newUser)
}

func getUser(context *gin.Context) {
	name := context.Param("name")
	userInput, err := getUserByName(name)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	context.IndentedJSON(http.StatusOK, userInput)
}

func updateUser(context *gin.Context) {
	name := context.Param("name")
	index, err := getUserByNameWithIndex(name)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	var updatedUser userInput
	if err := context.BindJSON(&updatedUser); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}

	// Update the user record with the new data
	userInputs[index] = updatedUser

	context.IndentedJSON(http.StatusOK, updatedUser)
}

func deleteUser(context *gin.Context) {
	name := context.Param("name")
	index, err := getUserByNameWithIndex(name)

	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	// Remove the user from the slice
	userInputs = append(userInputs[:index], userInputs[index+1:]...)

	context.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted"})
}

//getUserByName and getUserByNameWithIndex are both used as helper functions to navigate within the storage

func getUserByName(name string) (*userInput, error) {
	for i, t := range userInputs {
		if t.Name == name {
			return &userInputs[i], nil
		}
	}
	return nil, errors.New("user not found")
}

func getUserByNameWithIndex(name string) (int, error) {
	for i, t := range userInputs {
		if t.Name == name {
			return i, nil
		}
	}
	return -1, errors.New("user not found")
}

// These functions are used for the google sheets API
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
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
	defer f.Close()
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
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// functions for interacting with google sheets
func addUsertoGoogleSheets(user userInput) error {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := "10-CfbfktbeTSMV3tgnIKwaBquzw-RmjS13Tut9A32_s"
	writeRange := "Sheet1"

	var values [][]interface{}
	values = append(values, []interface{}{user.Name, user.Age, user.CommuteMethod, user.College, user.Hobbies})

	vr := &sheets.ValueRange{
		Values: values,
	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, vr).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	//initialize the router using gin
	router := gin.Default()

	router.GET("/getuser", getUsers)
	router.GET("/getuser/:name", getUser)
	router.PATCH("/updateuser/:name", updateUser)
	router.DELETE("/deleteuser/:name", deleteUser)
	router.POST("/adduser", addUser)
	router.Run("localhost:4000")

}
