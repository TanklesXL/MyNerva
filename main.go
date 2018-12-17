package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

const minerva = "https://horizon.mcgill.ca/pban1/twbkwbis.P_WWWLogin"

func main() {
	username, password, _ := credentials()
	fmt.Println(username, password)
}
func credentials() (username, password, phone string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ = reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password = string(bytePassword)

	// fmt.Print("\nEnter Phone Number: ")
	// phone, _ = reader.ReadString('\n')
	// phone = strings.Replace(strings.Replace(strings.Replace(strings.Replace(phone, " ", "", -1), "(", "", -1), ")", "", -1), "-", "", -1)

	return strings.TrimSpace(username), strings.TrimSpace(password), phone
}
