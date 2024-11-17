package misc

import (
	"fmt"
	"syscall"
	"time"

	"golang.org/x/term"
)

func GetTime() uint64 {
	return uint64(time.Now().Unix())
}

func ReadPassword(prompt string, verification bool) (string, error) {
	for {
		fmt.Print("Enter password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", err
		}
		password := string(bytePassword)
		fmt.Println()

		if verification {
			fmt.Print("Confirm password: ")
			bytePasswordConfirm, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return "", err
			}
			confirmPassword := string(bytePasswordConfirm)
			fmt.Println()
			if password == confirmPassword {
				return password, nil
			}
			fmt.Println("Passwords do not match. Please try again.")
		} else {
			return password, nil
		}
	}
}
