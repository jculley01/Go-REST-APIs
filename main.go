package main

import (
	"errors"
	"net/http"

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

func main() {
	router := gin.Default()
	router.GET("/getuser", getUsers)
	router.GET("/getuser/:name", getUser)
	router.PATCH("/updateuser/:name", updateUser)
	router.DELETE("/deleteuser/:name", deleteUser)
	router.POST("/adduser", addUser)
	router.Run("localhost:4000")

}
