package general_util

import (
	"fmt"
)

func GetUserInput() string {
	var input string
	fmt.Scanln(&input)
	return input
}

func PromptUserInput(prompt string) string {
	fmt.Print(prompt + ": ")
	userInput := GetUserInput()
	fmt.Print("\n")
	return userInput
}
