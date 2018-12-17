package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/chromedp/chromedp"

	"golang.org/x/crypto/ssh/terminal"
)

const minerva = "https://horizon.mcgill.ca/pban1/twbkwbis.P_WWWLogin"
const transcript = "https://horizon.mcgill.ca/pban1/bzsktran.P_Display_Form?user_type=S&tran_type=V"

var user, pass, phone, twilioSID, twilioToken string

func main() {
	//get user credentials
	credentials()

	//get the current transcript
	var oldTable string
	getTranscript(&oldTable)
	newTable := oldTable
	oldCourses := getCourses(oldTable)
	newCourses := getCourses(newTable)

	for {
		// every 10 minutes, check the transcript
		time.Sleep(10 * time.Minute)
		getTranscript(&newTable)
		newCourses = getCourses(newTable)
		//check if different and handle
		for key, newVal := range newCourses {
			if oldVal, ok := oldCourses[key]; !ok {
				notifyOnCourse(newVal)
			} else if oldVal != newVal {
				notifyOnCourse(newVal)
			}
		}
		//set the new one as the old one and try again
		oldTable = newTable
	}

}

func check(err error) {
	if err != nil {
		log.Fatal(err)
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

	// fmt.Print("\nEnter Phone Number: ")
	// phone, _ = reader.ReadString('\n')
	// phone = strings.Replace(strings.Replace(strings.Replace(strings.Replace(phone, " ", "", -1), "(", "", -1), ")", "", -1), "-", "", -1)

	// fmt.Print("Enter Twilo SID: ")
	// twilioSID, _ = reader.ReadString('\n')
	// twilioSID = strings.TrimSpace(strings.TrimSuffix(twilioSID, "\n"))

	// fmt.Print("Enter Twilio Auth Token: ")
	// twilioToken, _ = reader.ReadString('\n')
	// twilioToken = strings.TrimSpace(strings.TrimSuffix(twilioToken, "\n"))
}

func getCourses(table string) map[string]course {
	courses := make(map[string]course)

	reader := strings.NewReader(table)

	doc, err := goquery.NewDocumentFromReader(reader)
	check(err)

	doc.Find(`tr`).Each(func(i int, s1 *goquery.Selection) {
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

func getTranscript(table *string) {

	//create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create chrome instance
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
	check(err)

	// run task list
	tasks := chromedp.Tasks{
		//log-in
		chromedp.Navigate(minerva),
		chromedp.SendKeys(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="sid"]`, user),
		chromedp.SendKeys(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="PIN"]`, pass),
		chromedp.Submit(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="PIN"]`),
		chromedp.WaitNotPresent(`//form[@name="loginform1"]/table/tbody/tr/td[2]/input[@name="PIN"]`),
		//Go to transcript page
		chromedp.Navigate(transcript),
		chromedp.InnerHTML(`//body/div[3]/table[2]`, table),
		chromedp.Click(`//span[@class="pageheaderlinks"]/a[@accesskey="3"]`),
	}

	check(c.Run(ctxt, tasks))

	// shutdown chrome
	check(c.Shutdown(ctxt))

	// wait for chrome to finish
	check(c.Wait())
}

func notifyOnCourse(c course) {

}

type course struct {
	courseCode   string
	courseName   string
	yourMark     string
	classAverage string
}
