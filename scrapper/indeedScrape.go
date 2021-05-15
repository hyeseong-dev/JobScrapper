package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type job struct {
	id       string
	title    string
	company  string
	location string
	salary   string
	summary  string
}

// Scrape : scrape jobs from indeed.com
func IndeedScrape(term string) {
	var baseURL string = "https://kr.indeed.com/취업?as_and=" + term + "&radius=25&l=seoul&limit=50"
	startTime := time.Now()
	var allJobs []job
	c := make(chan []job)

	totalPages := getLastPages(term)

	for page := 0; page < totalPages; page++ {
		go getPage(page, baseURL, c)
	}

	for i := 0; i < totalPages; i++ {
		jobsInPage := <-c
		allJobs = append(allJobs, jobsInPage...)

	}

	writeJobs(allJobs)
	fmt.Println("Done, extracted", len(allJobs), "jobs from indeed.com")
	endTime := time.Now()
	fmt.Println("Operating time: ", endTime.Sub(startTime))
}

func writeJobs(allJobs []job) {
	c := make(chan []string)
	file, err := os.Create("indeed_jobs.csv")
	checkError(err)
	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"Title", "Company", "Location", "Salary", "Link", "Summary"}
	writeErr := w.Write(headers)
	checkError(writeErr)

	for _, job := range allJobs {
		go writeJobDetail(job, c)
	}

	for i := 0; i < len(allJobs); i++ {
		jobData := <-c
		writeErr := w.Write(jobData)
		checkError(writeErr)
	}
}

func writeJobDetail(job job, c chan<- []string) {
	const jobURL = "https://kr.indeed.com/viewjob?jk="
	c <- []string{job.title, job.company, job.location, job.salary, jobURL + job.id, job.summary}
}

func getPage(page int, url string, upperC chan<- []job) {
	const classCard = ".jobsearch-SerpJobCard"
	const pageConnection = "&start="
	var jobsInPage []job
	c := make(chan job)
	pageURL := url + pageConnection + strconv.Itoa(page*50)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkError(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})
	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobsInPage = append(jobsInPage, job)
	}
	upperC <- jobsInPage
}

func extractJob(card *goquery.Selection, c chan<- job) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	company := CleanString(card.Find(".company").Text())
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())
	c <- job{
		id:       id,
		title:    title,
		company:  company,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

// Remove all white spaces of head and tail of string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

// Get how many pages from the indeed website

// Get how many pages from the indeed website
func getLastPages(term string) int {
	var lastURL string = "https://kr.indeed.com/jobs?q=python&limit=50&start=9999"
	res, err := http.Get(lastURL)
	checkError(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	rawStr := strings.TrimSpace(doc.Find("#searchCountPages").Text())
	idx := strings.Index(rawStr, "페이지")
	lastPage, _ := strconv.Atoi(rawStr[:idx])
	return lastPage
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("request failed with Status: ", res.StatusCode)
	}
}
