package main

import (
	"context"
	"fmt"
	"github.com/sieglu2/virtual-friends-brain/llm"
)

func main() {
	ctx := context.Background()

	// Test creating a thread
	threadID, err := llm.CreateThreadWithFlask(ctx)
	if err != nil {
		fmt.Println("Error creating thread:", err)
		return
	}
	fmt.Println("Thread created successfully with ID:", threadID)

	// Set your assistant ID and content for testing
	assistantID := "asst_xIHAFLR0eWlTYrRcIeoG0xvj" // Replace with your actual assistant ID
	content := "Hello, what's your name?"

	// Test creating a message and running a thread
	messageResponse, err := llm.CreateMessageAndRunThreadWithFlask(ctx, threadID, assistantID, content)
	if err != nil {
		fmt.Println("Error creating message and running thread:", err)
		return
	}
	fmt.Println("Message response:", messageResponse)

	content = "what is my first question?"

	// Test creating a message and running a thread
	messageResponse, err = llm.CreateMessageAndRunThreadWithFlask(ctx, threadID, assistantID, content)
	if err != nil {
		fmt.Println("Error creating message and running thread:", err)
		return
	}
	fmt.Println("Message response:", messageResponse)

}
