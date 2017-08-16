package main

import (
	"fmt"
	"goSpace/goSpace"
)

func main() {

	// Create a space S
	inbox := goSpace.NewSpace("8080")

	// Put a message in the space
	goSpace.Put(inbox, "Hello World!")

	// Retrieve a message from the space
	var message string
	goSpace.Get(inbox, &message)

	// Print the message
	fmt.Println(message)
}
