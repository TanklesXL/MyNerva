package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/chromedp/chromedp"

	"golang.org/x/crypto/ssh/terminal"
)

const minerva = "https://horizon.mcgill.ca/pban1/twbkwbis.P_WWWLogin"
const transcript = "https://horizon.mcgill.ca/pban1/bzsktran.P_Display_Form?user_type=S&tran_type=V"

var user, pass string

func main() {
	user, pass, _ = credentials()

	//create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}

	// run task list
	var res string
	err = c.Run(ctxt, getTranscript(&res))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("got: `%s`", res)

	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

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

func getTranscript(res *string) chromedp.Tasks {
	return chromedp.Tasks{
		//log-in
		chromedp.Navigate(minerva),
		chromedp.SendKeys(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="sid"]`, user),
		chromedp.SendKeys(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="PIN"]`, pass),
		chromedp.Submit(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="PIN"]`),
		//chromedp.WaitNotVisible(`//input[@name="PIN"]`),

		//Go to transcript page
		//chromedp.Navigate(transcript),

		//chromedp.InnerHTML(`//div[@class="pagebodydiv]/table[1]`, res),
	}
}
