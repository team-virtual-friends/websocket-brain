package main

import (
	"fmt"
	"time"

	"github.com/pemistahl/lingua-go"
)

func main() {
	// info := getlang.FromString("hi, I am number 1.")
	// fmt.Println(info.LanguageCode(), info.LanguageName(), info.Confidence())
	// fmt.Printf("%+v\n", info)

	text := "hi, I am number 1."

	detector := lingua.NewLanguageDetectorBuilder().FromAllLanguages().
		Build()

	now := time.Now()
	// for i := 0; i < 10000; i++ {
	if language, exists := detector.DetectLanguageOf(text); exists {
		fmt.Println(language.IsoCode639_1().String())
		fmt.Println(language.IsoCode639_3().String())
	}
	// }
	fmt.Println(time.Now().Sub(now))
}
