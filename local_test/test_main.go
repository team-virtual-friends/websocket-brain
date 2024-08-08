package main

import (
	"fmt"
	"strings"
)

var (
	splitChars = []rune{'.', ';', '!', '?', ':', ',', '。', '；', '！', '？', '：', '，', '،', '।', '።', '။', '．'}
)

func main2() {
	testRun("是的，我xyz.可以說中文。請問a有什麼我，可以為您服務")
	testRun("服務")
	testRun("，")
	testRun("")
	// inputText := "的，"

}

func testRun(inputText string) {
	x, y := splitString(inputText)
	fmt.Println(x)
	fmt.Println(y)
	fmt.Println("----------")
}

func isSplitChar(char rune) bool {
	// Define a list of Unicode characters that represent a period in different languages
	periods := []rune{'.', ';', '!', '?', ':', ',', '。', '；', '！', '？', '：', '，', '،', '।', '።', '။', '．'}

	// Check if the character is in the list of periods
	for _, p := range periods {
		if char == p {
			return true
		}
	}

	return false
}

func splitString(text string) (string, string) {
	spliters := []rune{}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		res := isSplitChar(r)
		if res {
			spliters = append(spliters, r)
		}
		return res
	})
	// for _, x := range parts {
	// 	fmt.Println(x)
	// }
	// for i := 0; i < len(spliters); i++ {
	// 	fmt.Println(rune(spliters[i]))
	// }
	if len(parts) > 1 {
		first := strings.Builder{}
		second := strings.Builder{}
		i := 0
		for ; i < len(spliters); i++ {
			first.WriteString(parts[i])
			first.WriteRune(spliters[i])
		}
		for ; i < len(parts); i++ {
			second.WriteString(parts[i])
		}
		return first.String(), second.String()
	}
	return text, ""
}
