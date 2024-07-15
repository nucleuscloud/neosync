package main

import (
	"encoding/json"
	"fmt"
)

// Define the struct
type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Email     string `json:"email"`
}

func main() {
	// Create an instance of the struct
	person := Person{}

	// Convert the struct to JSON
	personJSON, err := json.Marshal(person)
	if err != nil {
		fmt.Println("Error marshalling struct to JSON:", err)
		return
	}

	// Print the JSON string
	fmt.Println(string(personJSON))
}
