package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/headzoo/surf"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/crypto/ssh/terminal"
)

const minerva = "https://horizon.mcgill.ca/pban1/twbkwbis.P_WWWLogin"
const transcript = "https://horizon.mcgill.ca/pban1/bzsktran.P_Display_Form?user_type=S&tran_type=V"
const logout = "https://horizon.mcgill.ca/pban1/twbkwbis.P_Logout"

var user, pass, phone, twilPhone, twilioSID, twilioToken, twilio string

func main() {
	//get user credentials
	credentials()

	//get the current transcript
	var oldTable, newTable *goquery.Selection
	var oldCourses, newCourses map[string]course
	oldTable = getTranscriptWithSurf()
	oldCourses = getCourses(oldTable)
	twilio = "https://api.twilio.com/2010-04-01/Accounts/" + twilioSID + "/Messages.json"
	notify("\nConfirmation of phone number for MyNerva.")

	for {
		newTable = getTranscriptWithSurf()
		newCourses = getCourses(newTable)
		//check if different and handle
		for key, newVal := range newCourses {
			if oldVal, ok := oldCourses[key]; !ok {
				notify(newVal.constructMessage())
			} else if oldVal != newVal {
				notify(newVal.constructMessage())
			}
		}
		//set the new one as the old one and try again
		oldTable = newTable
		time.Sleep(10 * time.Minute)
	}

}

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

func credentials() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')
	user = strings.TrimSpace(strings.TrimSuffix(username, "\n"))

	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	pass = strings.TrimSpace(strings.TrimSuffix(string(bytePassword), "\n"))

	fmt.Print("\nEnter Destination Phone Number: ")
	phone, _ = reader.ReadString('\n')
	phone = "+1" + strings.Replace(strings.Replace(strings.Replace(strings.Replace(phone, " ", "", -1), "(", "", -1), ")", "", -1), "-", "", -1)

	fmt.Print("Enter Twilio Phone Number: ")
	twilPhone, _ = reader.ReadString('\n')
	twilPhone = "+1" + strings.Replace(strings.Replace(strings.Replace(strings.Replace(twilPhone, " ", "", -1), "(", "", -1), ")", "", -1), "-", "", -1)

	fmt.Print("Enter Twilo SID: ")
	twilioSID, _ = reader.ReadString('\n')
	twilioSID = strings.TrimSpace(strings.TrimSuffix(twilioSID, "\n"))

	fmt.Print("Enter Twilio Auth Token: ")
	twilioToken, _ = reader.ReadString('\n')
	twilioToken = strings.TrimSpace(strings.TrimSuffix(twilioToken, "\n"))
}

func getCourses(table *goquery.Selection) map[string]course {
	courses := make(map[string]course)
	table.Find(`tr`).Each(func(i int, s1 *goquery.Selection) {
		tdTags := s1.Find(`td`)
		if tdTags.Length() == 11 {
			var c course
			tdTags.Each(func(j int, s2 *goquery.Selection) {
				switch s2.Index() {
				case 1:
					c.courseCode = strings.TrimSpace(s2.Text())
				case 3:
					c.courseName = strings.TrimSpace(s2.Text())
				case 6:
					c.yourMark = strings.TrimSpace(s2.Text())
				case 10:
					c.classAverage = strings.TrimSpace(s2.Text())
				}
			})
			courses[c.courseCode] = c
		}
	})

	return courses
}

func getTranscriptWithSurf() *goquery.Selection {
	bow := surf.NewBrowser()
	bow.Open(minerva)
	fm, _ := bow.Form(`form[name="loginform1"]`)
	fm.Input("sid", user)
	fm.Input("PIN", pass)
	fm.Submit()
	bow.Open(transcript)
	if strings.TrimSpace(bow.Title()) != "UNOFFICIAL Transcript for ID" {
		fmt.Println("\nLOGIN FAILED")
		os.Exit(0)
	}
	outputSel := bow.Find(`table.dataentrytable`).Last()
	check(bow.Open(logout))
	return outputSel
}

func notify(message string) {
	msgData := url.Values{}
	msgData.Set("To", phone)
	msgData.Set("From", twilPhone)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())
	client := &http.Client{}
	req, _ := http.NewRequest("POST", twilio, &msgDataReader)
	req.SetBasicAuth(twilioSID, twilioToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println("Message Sent: " + message)
		}
	} else {
		fmt.Println(resp.Status)
	}
}

type course struct {
	courseCode   string
	courseName   string
	yourMark     string
	classAverage string
}

type message string

func (m message) constructMessage() string {
	return string(m)
}
func (c course) constructMessage() string {
	var sb strings.Builder
	sb.WriteString("Grade Notification\n")
	sb.WriteString(c.courseCode + "\n")
	sb.WriteString("Your Grade: " + c.yourMark + "\n")
	sb.WriteString("Class Average: " + c.classAverage + "\n")
	return sb.String()
}
