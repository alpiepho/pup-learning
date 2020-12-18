package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// MAXCOUNT -
const MAXCOUNT int = 2000

// URLBase -
const URLBase string = "https://www.linkedin.com"

// URLLogin -
const URLLogin string = "https://www.linkedin.com/learning/me"

// URLLogout -
const URLLogout string = "https://www.linkedin.com/learning/logout"

// URLHistory -
const URLHistory string = "https://www.linkedin.com/learning/me/completed?trk=nav_neptune_learning"

// Course is course
type Course struct {
	title         string
	link          string
	author        string
	releasedDate  string
	duration      string
	completedDate string
	img           string
	imgFile       string
	linkedin      string
	details       string
	completedTs   int64
}

// ByCompleted -
type ByCompleted []Course

func (a ByCompleted) Len() int           { return len(a) }
func (a ByCompleted) Less(i, j int) bool { return a[i].completedTs > a[j].completedTs }
func (a ByCompleted) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func loginAuto(ctx *context.Context) {
	err := chromedp.Run(*ctx,
		chromedp.Navigate(URLLogin),
		chromedp.SendKeys(`#auth-id-input`, os.Getenv("LI_USERNAME"), chromedp.ByID),
		chromedp.Sleep(2*time.Second),
		chromedp.Click(`#auth-id-button`, chromedp.ByID),
		chromedp.Sleep(2*time.Second),
		chromedp.SendKeys(`#password`, os.Getenv("LI_PASSWORD"), chromedp.ByID),
		chromedp.Sleep(2*time.Second),
		chromedp.Click(`button.btn__primary--large`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// // loginManual in case site starts asking questions
// func loginManual(ctx *context.Context) {
// 	err := chromedp.Run(*ctx,
// 		chromedp.Navigate(URLLogin),
// 	)
// 	if err != nil {
// 		// ignore error
// 		log.Fatal(err)
// 	}
// 	fmt.Println(os.Getenv("LI_USERNAME"))
// 	fmt.Println(os.Getenv("LI_PASSWORD"))
// 	fmt.Println("FINISH LOGIN within 60 seconds...")
// 	time.Sleep(60 * time.Second)
// }

func logoutAuto(ctx *context.Context) {
	err := chromedp.Run(*ctx,
		chromedp.Navigate(URLLogout),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		// ignore error
	}
}

func autoScroll(ctx *context.Context) {
	totalHeight := 0
	scrollHeight := 1
	for totalHeight < scrollHeight {
		// fmt.Printf("%d\n", totalHeight)
		// fmt.Printf("%d\n", scrollHeight)
		totalHeight += 400
		err := chromedp.Run(*ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				_, exp, err := runtime.Evaluate(`window.scrollBy(0,400);`).Do(ctx)
				if err != nil {
					return err
				}
				if exp != nil {
					return exp
				}
				return nil
			}),
			chromedp.Sleep(2*time.Second),
			chromedp.EvaluateAsDevTools(`document.body.scrollHeight`, &scrollHeight),
		)
		if err != nil {
			// ignore error
		}
	}
}

func fixDate(input string) string {
	parts := strings.Split(input, "/")
	if len(parts) == 3 {
		m, _ := strconv.Atoi(parts[0])
		d, _ := strconv.Atoi(parts[1])
		y, _ := strconv.Atoi(parts[2])
		input = fmt.Sprintf("%02d/%02d/%04d", m, d, y)
	}
	return input
}

func parseHistory(ctx *context.Context, courses *[]Course, noscroll bool) {
	var err error
	err = chromedp.Run(*ctx,
		chromedp.Navigate(URLHistory),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		// ignore error
	}

	if !noscroll {
		autoScroll(ctx)
	}

	var nodes []*cdp.Node
	err = chromedp.Run(*ctx,
		chromedp.Nodes(`.lls-card-detail-card`, &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		// ignore error
	}

	fmt.Println("Parsing courses")
	fmt.Println(len(nodes))
	for i, n := range nodes {
		var ok bool
		data := Course{}
		err := chromedp.Run(*ctx,
			chromedp.Text(`h3`, &data.title, chromedp.ByQuery, chromedp.FromNode(n)),
			chromedp.AttributeValue(`h3 a`, "href", &data.link, &ok, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			chromedp.Text(`.lls-card-authors`, &data.author, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			chromedp.Text(`.lls-card-released-on`, &data.releasedDate, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			chromedp.Text(`.lls-card-duration-label`, &data.duration, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			chromedp.Text(`.lls-card-completion-state--completed`, &data.completedDate, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			chromedp.AttributeValue(`img`, "src", &data.img, &ok, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			chromedp.AttributeValue(`.lls-card-authors a`, "href", &data.linkedin, &ok, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
		)
		if err != nil {
			// ignore error
		}
		fmt.Println(i)
		fmt.Println(data.title)
		data.link = URLBase + data.link
		data.author = strings.Replace(data.author, "By: ", "", 1)
		data.releasedDate = strings.Replace(data.releasedDate, "Released ", "", 1)
		data.releasedDate = strings.Replace(data.releasedDate, "Updated ", "", 1)
		data.completedDate = strings.Replace(data.completedDate, "Completed ", "", 1)
		// REFERENCE: https://programming.guide/go/format-parse-string-time-date-example.html
		t, _ := time.Parse("01/02/2006", fixDate(data.completedDate))
		data.completedTs = t.Unix()
		data.imgFile = ""
		data.linkedin = URLBase + data.linkedin
		*courses = append(*courses, data)
	}
	fmt.Println("Parsing courses, done.")
	fmt.Println(len(nodes))
}

func saveThumbs(ctx *context.Context, courses *[]Course, nopngs bool) {
	var err error
	err = os.RemoveAll("./public/images")
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("./public/images", 0755)
	if err != nil {
		log.Fatal(err)
	}
	if nopngs {
		return
	}
	for i, course := range *courses {
		if len(course.img) > 0 {
			parts := strings.Split(course.img, "/")
			course.imgFile = "./public/images/" + parts[5] + ".png"
			(*courses)[i].imgFile = "./images/" + parts[5] + ".png"

			response, e := http.Get(course.img)
			if e != nil {
				log.Fatal(e)
			}
			defer response.Body.Close()

			file, err := os.Create(course.imgFile)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			_, err = io.Copy(file, response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func parseDetails(ctx *context.Context, courses []Course, getexfiles bool) {
	for _, course := range courses {
		var details string = ""
		err := chromedp.Run(*ctx,
			chromedp.Navigate(course.link),
			chromedp.Text(`p.section-container__description`, &details, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
		)
		if err != nil {
			// ignore error
		}
		course.details = details
		if getexfiles {
			time.Sleep(2 * time.Second)
			err = chromedp.Run(*ctx,
				chromedp.Click(`button[aria-label="Show all exercise files"]`, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
				chromedp.Sleep(2*time.Second),
			)
			if err == nil {
				err = chromedp.Run(*ctx,
					chromedp.Click(`button[aria-label="Dismiss"]`, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
					chromedp.Sleep(2*time.Second),
				)
				if err == nil {
				}
			}

			// <button aria-label="Show all exercise files" id="ember437" class="artdeco-button artdeco-button--1 artdeco-button--tertiary ember-view btn-link" data-control-name="exercise_files_modal"><!---->
			// <span class="artdeco-button__text">
			// Show all
			// </span></button>

			// <a tabindex="0" rel="noopener noreferrer" target="_blank" href="https://files3.lynda.com/secure/courses/696863/exercises/Ex_Files_Selenium_EssT.zip?6fTMprnWbUDxx-qEigGE75YmeHtaOYpfbqgIq9-P5qDNMvt_wyqTDCeZey7enPUBmguuYc0T_4-ITcXWWCc0lZx9P9HKtKCSeNsgBhi5eydEDoyB1JmHPSdPXYbqaH7fd1to-AgLfP8VQZcH5B7XSjcDZLTea9cboQ" id="ember530" class="ember-view classroom-exercise-files-modal__exercise-file-download artdeco-button artdeco-button--secondary">
			// 		<span aria-hidden="true">
			// 		Download
			// 	</span>
			// 	<span class="a11y-text">
			// 		Download Ex_Files_Selenium_EssT.zip with file size 1.7MB
			// 	</span>
			// </a>

			// 			<button data-test-modal-close-btn="" aria-label="Dismiss" id="ember533" class="artdeco-modal__dismiss artdeco-button artdeco-button--circle artdeco-button--muted artdeco-button--2 artdeco-button--tertiary ember-view">  <li-icon aria-hidden="true" type="cancel-icon" class="artdeco-button__icon"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" data-supported-dps="24x24" fill="currentColor" width="24" height="24" focusable="false">
			//   <path d="M20 5.32L13.32 12 20 18.68 18.66 20 12 13.33 5.34 20 4 18.68 10.68 12 4 5.32 5.32 4 12 10.69 18.68 4z"></path>
			// </svg></li-icon>

			// <span class="artdeco-button__text">

			// </span></button>

		}
	}
}

func buildTimes(courses []Course) (int, int) {
	// Derive timestamps and duration, sort
	totalSec := 0
	for _, course := range courses {
		// assume "An Bm" or "Bm"
		parts := strings.Split(course.duration, " ")
		totalSec++
		for _, part := range parts {
			if strings.Contains(part, "h") {
				val, _ := strconv.Atoi(strings.Replace(part, "h", "", 1))
				totalSec += val * 60 * 60
			}
			if strings.Contains(part, "m") {
				val, _ := strconv.Atoi(strings.Replace(part, "m", "", 1))
				totalSec += val * 60
			}
			if strings.Contains(part, "s") {
				val, _ := strconv.Atoi(strings.Replace(part, "s", "", 1))
				totalSec += val
			}
		}
	}
	totalMin := totalSec / 60
	totalH := totalMin / 60
	totalM := totalMin - (totalH * 60)
	return totalH, totalM
}

func writeToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

// HTML1 -
const HTML1 string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width", initial-scale=1.0"/>
    <meta name="Description" content="LinkedIn Learning Courses Completed">
    <meta name="theme-color" content="#d36060"/>
    <title>
    LinkedIn Learning Courses Completed
    </title>
    <link rel="stylesheet" href="./style.css" />
    <link rel="manifest" href="./manifest.json" />
    <link rel="icon"
      type="image/png" 
      href="./favicon.ico" />
  </head>
  <body class="body">
    <main>
    <article class="page">
      <h1  id="top">LinkedIn Learning Courses Completed</h1>

      <div class="introduction">
      <p>
      This a summary of all the Linked-In courses I have completed. 
      This is just the direct LinkedIn Learning courses.  There are a number of "Lynda.com"
      courses that were taken before subscribing to the LinkedIn premium plan.
      </p>
      <p>
      This list is generated from a tool called "main.go" that can be found
      <a
        href="https://github.com/alpiepho/pup-learning"
        target="_blank"
        rel="noreferrer"
      >here</a>.  This tool needs to be run manually to parse the LinkedIn Learning
      site to gather the list of courses I have taken.
      </p>
      <p>
        If you look over the list of courses, there is variety.  I fully admit that
        my attention for some courses was less that other.  My form of bing watching :) 
      </p>
      </div>
`

// HTML2 -
const HTML2 string = `
    <div id="bottom"></div>
    </article>
  </body>
</html>
`

func buildHTML(courses []Course, totalH int, totalM int) {
	today := time.Now().Local()
	var b strings.Builder
	fmt.Fprintf(&b, "%s", HTML1)
	fmt.Fprintf(&b, "<sup><sub>(updated %s)</sub></sup>\n\n", today)
	fmt.Fprintf(&b, "      <br/><p>Totals - Course: %d, Time: %dh %dm</p><br/>\n\n", len(courses), totalH, totalM)
	fmt.Fprintf(&b, "      <ul class=\"mainul\">\n")
	for _, course := range courses {
		fmt.Fprintf(&b, "            <li>\n")
		fmt.Fprintf(&b, "              <ul>\n")
		fmt.Fprintf(&b, "                <li>\n")
		fmt.Fprintf(&b, "                  <div class=\"leftside\">\n")
		if len(course.imgFile) > 0 {
			fmt.Fprintf(&b, "                    <p><img src=\"%s\" loading=\"lazy\"</img></p>\n", course.imgFile)
		} else {
			fmt.Fprintf(&b, "                    <p><img src=\"%s\" loading=\"lazy\"</img></p>\n", course.img)
		}
		fmt.Fprintf(&b, "                  </div>\n")
		fmt.Fprintf(&b, "                  <div class=\"rightside\">\n")
		fmt.Fprintf(&b, "                    <ul>\n")
		fmt.Fprintf(&b, "                      <li>\n")
		fmt.Fprintf(&b, "                        <a target=\"_blank\" href=\"%s\">\n", course.link)
		fmt.Fprintf(&b, "                        %s\n", course.title)
		fmt.Fprintf(&b, "                        </a>  ")
		fmt.Fprintf(&b, "                      </li>\n")
		fmt.Fprintf(&b, "                      <li>\n")
		fmt.Fprintf(&b, "                        <span>(%s ... %s)</span>\n", course.releasedDate, course.duration)
		fmt.Fprintf(&b, "                      </li>\n")
		fmt.Fprintf(&b, "                      <li>\n")
		if len(course.linkedin) > 0 {
			fmt.Fprintf(&b, "                        <li>Author: <a target=\"_blank\" href=\"%s\">%s</a></li>\n", course.linkedin, course.author)
		} else {
			fmt.Fprintf(&b, "                        <li>Author: %s</li>\n", course.author)
		}
		fmt.Fprintf(&b, "                      </li>\n")
		fmt.Fprintf(&b, "                      <li>\n")
		fmt.Fprintf(&b, "                        <li class=\"completed\"><i>Completed: %s</i></li>\n", course.completedDate)
		fmt.Fprintf(&b, "                      </li>\n")
		fmt.Fprintf(&b, "                    </ul>\n")
		fmt.Fprintf(&b, "                  </div>\n")
		fmt.Fprintf(&b, "                </li>\n")
		fmt.Fprintf(&b, "                <li class=\"details\">\n")
		fmt.Fprintf(&b, "                  %s\n", course.details)
		fmt.Fprintf(&b, "                </li>\n")
		fmt.Fprintf(&b, "                <li class=\"topbottom\"><a href=\"#top\">top</a> / <a href=\"#bottom\">bottom</a></li>\n")
		fmt.Fprintf(&b, "              </ul>\n")
		fmt.Fprintf(&b, "            </li>\n")
	}
	fmt.Fprintf(&b, "      </ul>\n")
	fmt.Fprintf(&b, "%s", HTML2)

	err := writeToFile("./public/index.html", b.String())
	if err != nil {
		// ignore error
	}
}

func main() {
	noscroll := flag.Bool("noscroll", false, "a bool")
	nopngs := flag.Bool("nopngs", false, "a bool")
	nosort := flag.Bool("nosort", false, "a bool")
	getexfiles := flag.Bool("getexfiles", false, "a bool")
	flag.Parse()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", os.Getenv("GO_HEADLESS") == "true"),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var courses []Course
	loginAuto(&ctx)
	parseHistory(&ctx, &courses, *noscroll)
	saveThumbs(&ctx, &courses, *nopngs)
	if *getexfiles {
		parseDetails(&ctx, courses, *getexfiles)
		logoutAuto(&ctx)
	} else {
		logoutAuto(&ctx)
		parseDetails(&ctx, courses, *getexfiles)
	}
	if !*nosort {
		sort.Sort(ByCompleted(courses))
	}
	totalH, totalM := buildTimes(courses)

	buildHTML(courses, totalH, totalM)
}
