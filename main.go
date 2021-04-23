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
const URLLogout string = "https://www.linkedin.com/uas/logout"

//https://www.linkedin.com/uas/logout/?csrfToken=ajax%3A6635960707050209149&session_redirect=https%3A%2F%2Fwww.linkedin.com%2Flearning%2F&lipi=urn%3Ali%3Apage%3Ad_learning_me_history%3B%2BSuYFugwSvOAgXJ9cvsvVg%3D%3D&licu

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

// loginManual in case site starts asking questions
func loginManual(ctx *context.Context) {
	err := chromedp.Run(*ctx,
		chromedp.Navigate(URLLogin),
	)
	if err != nil {
		// ignore error
		log.Fatal(err)
	}
	fmt.Println(os.Getenv("LI_USERNAME"))
	fmt.Println(os.Getenv("LI_PASSWORD"))
	fmt.Println("FINISH LOGIN within 60 seconds...")
	time.Sleep(60 * time.Second)
}

func logoutAuto(ctx *context.Context) {
	err := chromedp.Run(*ctx,
		chromedp.Navigate(URLLogout),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		// ignore error
	}
}

func doLogin(ctx *context.Context, manuallogin bool, stages bool, stagelogin bool) {
	fmt.Printf("stages %t\n", stages)
	fmt.Printf("stagelogin %t\n", stagelogin)

	if !stages || stagelogin {
		if manuallogin {
			loginManual(ctx)
		} else {
			loginAuto(ctx)
		}
	}
}

func doLogout(ctx *context.Context, stages bool, stagelogout bool) {
	if !stages || stagelogout {
		logoutAuto(ctx)
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

func parseHistory(ctx *context.Context, courses *[]Course, noscroll bool, stages bool, stagehistory bool) {
	if !stages || stagehistory {
		var err error
		err = chromedp.Run(*ctx,
			chromedp.Navigate(URLHistory),
			chromedp.Sleep(4*time.Second),
		)
		if err != nil {
			// ignore error
		}

		if !noscroll {
			autoScroll(ctx)
		}

		fmt.Println("Parsing Nodes...")
		var nodes []*cdp.Node
		err = chromedp.Run(*ctx,
			chromedp.Nodes(`.completed-body__card`, &nodes, chromedp.ByQueryAll),
		)
		if err != nil {
			// ignore error
			fmt.Println(err)
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
				chromedp.Text(`.lls-card-thumbnail-label`, &data.duration, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
				chromedp.Text(`.lls-card-completion-state--completed`, &data.completedDate, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
				chromedp.AttributeValue(`img`, "src", &data.img, &ok, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
				chromedp.AttributeValue(`.lls-card-authors a`, "href", &data.linkedin, &ok, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0), chromedp.FromNode(n)),
			)
			if err != nil {
				// ignore error
				fmt.Println(err)
			}
			//DEBUG
			fmt.Println(i)
			fmt.Println(data.title)
			// fmt.Println(data.link)
			// fmt.Println(data.author)
			// fmt.Println(data.releasedDate)
			// fmt.Println(data.duration)
			// fmt.Println(data.completedDate)
			// fmt.Println(data.img)
			// fmt.Println(data.linkedin)
			// fmt.Println(n)
			// fmt.Println(data)
			// fmt.Println("")

			data.link = URLBase + data.link
			data.author = strings.Replace(data.author, "By: ", "", 1)
			data.releasedDate = strings.Replace(data.releasedDate, "Released ", "", 1)
			data.releasedDate = strings.Replace(data.releasedDate, "Updated ", "", 1)
			data.completedDate = strings.Replace(data.completedDate, "Completed ", "", 1)

			//DEBUG
			//time.Sleep(6000 * time.Second)

			// REFERENCE: https://programming.guide/go/format-parse-string-time-date-example.html
			t, _ := time.Parse("01/02/2006", fixDate(data.completedDate))
			data.completedTs = t.Unix()
			data.imgFile = ""
			data.linkedin = URLBase + data.linkedin
			if len(data.completedDate) > 0 {
				*courses = append(*courses, data)
			}
		}
		fmt.Println("Parsing courses, done.")
		fmt.Println(len(nodes))
	}
}

func saveThumbs(ctx *context.Context, courses *[]Course, nopngs bool, stages bool, stagethumbs bool) {
	if !stages || stagethumbs {
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
}

func parseDetails(ctx *context.Context, courses *[]Course, getexfiles bool, stages bool, stagedetails bool) {
	if !stages || stagedetails {
		for i, course := range *courses {
			var details string = ""
			time.Sleep(2 * time.Second)
			//time.Sleep(4 * time.Second)
			err := chromedp.Run(*ctx,
				chromedp.Navigate(course.link),
				// logged out
				chromedp.Text(`p.section-container__description`, &details, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
				// logged in?
				//chromedp.Text(`p.t-16`, &details, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			)
			//time.Sleep(60 * time.Second)

			if err != nil {
				// ignore error
				fmt.Println(course.link)
				fmt.Println(err)
			}
			(*courses)[i].details = details
			if getexfiles {
				//time.Sleep(2 * time.Second)
				err = chromedp.Run(*ctx,
					chromedp.Click(`button[aria-label="Show all exercise files"]`, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
					chromedp.Sleep(2*time.Second),
				)
				if err == nil {
					err = chromedp.Run(*ctx,
						chromedp.Click(`a.classroom-exercise-files-modal__exercise-file-download`, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
						chromedp.Sleep(10*time.Second),
					)
					err = chromedp.Run(*ctx,
						chromedp.Click(`button[aria-label="Dismiss"]`, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
						chromedp.Sleep(2*time.Second),
					)
				}
			}
		}
	}
}

func buildTimes(courses []Course, stages bool, stagetimes bool) (int, int) {
	// Derive timestamps and duration, sort
	if !stages || stagetimes {
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
	return 0, 0
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

func buildHTML(courses []Course, totalH int, totalM int, stages bool, stagehtml bool) {
	if !stages || stagehtml {
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
}

func main() {
	manuallogin := flag.Bool("manuallogin", false, "a bool")
	stages := flag.Bool("stages", false, "a bool, enables stages -login, -history, -thumbs, -logout, -details, -times, -html")
	stagelogin := flag.Bool("login", false, "a bool")
	stagehistory := flag.Bool("history", false, "a bool")
	stagethumbs := flag.Bool("thumbs", false, "a bool")
	//stagelogout := flag.Bool("logout", false, "a bool")
	stagedetails := flag.Bool("details", false, "a bool")
	stagetimes := flag.Bool("times", false, "a bool")
	stagehtml := flag.Bool("html", false, "a bool")
	noscroll := flag.Bool("noscroll", false, "a bool")
	nopngs := flag.Bool("nopngs", false, "a bool")
	nosort := flag.Bool("nosort", false, "a bool")
	getexfiles := flag.Bool("getexfiles", false, "a bool")
	flag.Parse()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", os.Getenv("GO_HEADLESS") == "true"),
	)

	//DEBUG for running in debugger
	// fmt.Println("DEBUG hardcoded")
	// *manuallogin = false
	// *noscroll = true
	// *stages = true
	// *stagelogin = true
	// *stagehistory = true
	// *stagethumbs = true
	// *stagelogout = true
	// *stagedetails = true
	// *stagetimes = true
	// *stagehtml = true
	// fmt.Printf("manuallogin %t\n", *manuallogin)
	// fmt.Printf("noscroll %t\n", *noscroll)
	// fmt.Printf("stages %t\n", *stages)
	// fmt.Printf("stagelogin %t\n", *stagelogin)
	// fmt.Printf("stagehistory %t\n", *stagehistory)
	// fmt.Printf("stagethumbs %t\n", *stagethumbs)
	// fmt.Printf("stagelogout %t\n", *stagelogout)
	// fmt.Printf("stagedetails %t\n", *stagedetails)
	// fmt.Printf("stagetimes %t\n", *stagetimes)
	// fmt.Printf("stagehtml %t\n", *stagehtml)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var courses []Course
	doLogin(&ctx, *manuallogin, *stages, *stagelogin)
	parseHistory(&ctx, &courses, *noscroll, *stages, *stagehistory)
	saveThumbs(&ctx, &courses, *nopngs, *stages, *stagethumbs)
	if *getexfiles {
		parseDetails(&ctx, &courses, *getexfiles, *stages, *stagedetails)
		logoutAuto(&ctx)
	} else {
		// logout doesnt work (needs cookie) so try new browser
		//doLogout(&ctx, *stages, *stagelogout)
		cancel()
		ctx2, cancel2 := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel2()
		ctx2, cancel2 = chromedp.NewContext(ctx2)
		defer cancel2()
		parseDetails(&ctx2, &courses, *getexfiles, *stages, *stagedetails)
	}
	if !*nosort {
		sort.Sort(ByCompleted(courses))
	}
	totalH, totalM := buildTimes(courses, *stages, *stagetimes)

	buildHTML(courses, totalH, totalM, *stages, *stagehtml)
}
