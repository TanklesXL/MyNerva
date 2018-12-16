package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/crypto/ssh/terminal"
)

const minerva = "https://horizon.mcgill.ca/pban1/twbkwbis.P_WWWLogin"

func main() {

	username, password, _ := credentials()

	jar, _ := cookiejar.New(nil)

	app := App{
		Client: &http.Client{Jar: jar},
	}

	app.login(username, password)

}

// App holds the http client
type App struct {
	Client *http.Client
}

// AuthenticityToken contains the token to log into minerva
type AuthenticityToken struct {
	Token string
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

func (app *App) getToken() AuthenticityToken {
	client := app.Client

	response, err := client.Get(minerva)
	if err != nil {
		log.Fatalln("Error fetching response. ", err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	token, _ := document.Find("input[name='authenticity_token']").Attr("value")

	authenticityToken := AuthenticityToken{
		Token: token,
	}

	return authenticityToken
}

func (app *App) login(username, password string) {
	client := app.Client

	authenticityToken := app.getToken()

	data := url.Values{
		"authenticity_token": {authenticityToken.Token},
		"user[login]":        {username},
		"user[password]":     {password},
	}

	response, err := client.PostForm(minerva, data)

	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()

	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
}
