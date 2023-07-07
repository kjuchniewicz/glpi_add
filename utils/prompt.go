package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

func PromptString(name string) (string, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: ValidateEmptyInput,
	}

	return prompt.Run()
}

func PromptInteger(name string) (int64, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: ValidateIntegerNumberInput,
	}

	promptResult, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	parseInt, _ := strconv.ParseInt(promptResult, 0, 0)
	return parseInt, nil
}

func ValidateEmptyInput(input string) error {
	if len(strings.TrimSpace(input)) < 1 {
		return errors.New("nie może być pusto")
	}
	return nil
}

func ValidateIntegerNumberInput(input string) error {
	_, err := strconv.ParseInt(input, 0, 0)
	if err != nil {
		return errors.New("nieprawidłowy numer")
	}
	return nil
}
