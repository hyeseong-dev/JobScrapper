package main

import (
	"os"
	"strings"

	"github.com/hyeseong-dev/jobscrapper/scrapper"
	"github.com/labstack/echo"
)

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleIndeedScrape(c echo.Context) error {
	fileName := "indeed_jobs.csv"
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.IndeedScrape(term)
	return c.Attachment(fileName, fileName)
	return nil
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleIndeedScrape)
	e.Logger.Fatal(e.Start(":4000"))
}
