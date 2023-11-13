package main

import (
	"fmt"
	"unicode"
)

func main() {
	inputText := "This is a sentence. Another one? Yes, indeed!"

	for _, ch := range inputText {
		fmt.Println(string(ch), unicode.IsPunct(ch))
	}
}
