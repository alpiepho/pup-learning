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
		chromedp.Sleep(2*time.Second),
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
	retries := 0
	for totalHeight < scrollHeight && retries < 20 {
		fmt.Printf("retries      %d\n", retries)
		fmt.Printf("totalHeight  %d\n", totalHeight)
		fmt.Printf("scrollHeight %d\n", scrollHeight)
		totalHeight += 1600 //400
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
			retries += 1
		}
		if err == nil {
			retries = 0
		}
		// try More button
		err = chromedp.Run(*ctx,
			chromedp.Click(`button[aria-label="Show more learning history"]`, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			chromedp.Sleep(1*time.Second),
		)
		if err == nil {
			//totalHeight = 0
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

// LI-AI-LIST
var li_ai_links []string = []string{
	"https://www.linkedin.com/learning/a-practical-guide-to-upskilling-your-organization-on-ai",
	"https://www.linkedin.com/learning/recurrent-neural-networks",
	"https://www.linkedin.com/learning/32db3356-251f-39a4-9652-3326ac346d04",
	"https://www.linkedin.com/learning/ai-fundamentals-for-data-professionals",
	"https://www.linkedin.com/learning/how-to-boost-your-productivity-with-ai-tools",
	"https://www.linkedin.com/learning/deep-learning-getting-started",
	"https://www.linkedin.com/learning/machine-learning-and-artificial-intelligence-security-risk-categorizing-attacks-and-failure-modes",
	"https://www.linkedin.com/learning/deep-learning-model-optimization-and-tuning",
	"https://www.linkedin.com/learning/responsible-ai-principles-and-practical-applications-high-visibility",
	"https://www.linkedin.com/learning/ai-workshop-hands-on-with-gans-using-dense-neural-networks",
	"https://www.linkedin.com/learning/ai-workshop-build-a-neural-network-with-pytorch-lightning",
	"https://www.linkedin.com/learning/ai-workshop-hands-on-with-gans-with-deep-convolutional-networks",
	"https://www.linkedin.com/learning/training-neural-networks-in-python-17058600",
	"https://www.linkedin.com/learning/l-intelligence-artificielle-ia-generative-pour-les-dirigeants",
	"https://www.linkedin.com/learning/openai-api-building-assistants",
	"https://www.linkedin.com/learning/amplify-your-communication-skills-with-generative-ai",
	"https://www.linkedin.com/learning/building-a-responsible-ai-program-context-culture-content-commitment",
	"https://www.linkedin.com/learning/openai-api-code-interpreter-and-advanced-data-analysis",
	"https://www.linkedin.com/learning/openai-api-speech",
	"https://www.linkedin.com/learning/openai-api-embeddings",
	"https://www.linkedin.com/learning/ya-xu-how-to-turn-ai-from-a-buzz-word-to-a-business-tool",
	"https://www.linkedin.com/learning/introduction-to-auditing-ai-systems",
	"https://www.linkedin.com/learning/generative-ai-foundations-introduction-to-generative-adversarial-networks",
	"https://www.linkedin.com/learning/9306e21e-c229-3a39-a7bc-c7a45ec3ed96",
	"https://www.linkedin.com/learning/artificial-intelligence-foundations-neural-networks-22853427",
	"https://www.linkedin.com/learning/uretken-yz-nedir",
	"https://www.linkedin.com/learning/artificial-intelligence-for-cybersecurity-2023-revision",
	"https://www.linkedin.com/learning/stable-diffusion-tips-tricks-and-techniques-high-visibility",
	"https://www.linkedin.com/learning/ia-generative-et-creation-de-contenus-opportunites-risques-et-ethique",
	"https://www.linkedin.com/learning/generative-ai-imaging-what-creative-professionals-need-to-know",
	"https://www.linkedin.com/learning/how-to-research-and-write-using-generative-ai-tools",
	"https://www.linkedin.com/learning/prompt-engineering-how-to-talk-to-the-ais-high-visibility",
	"https://www.linkedin.com/learning/leveraging-ai-in-adobe-photoshop-and-creative-cloud",
	"https://www.linkedin.com/learning/midjourney-tips-and-techniques-for-creating-images",
	"https://www.linkedin.com/learning/introduction-to-large-language-models",
	"https://www.linkedin.com/learning/nano-tips-for-using-chat-gpt-for-business",
	"https://www.linkedin.com/learning/scaling-generative-ai-building-a-strategy-for-adoption-and-expansion-high-visibility",
	"https://www.linkedin.com/learning/ai-and-the-future-of-work-workflows-and-modern-tools-for-tech-leaders",
	"https://www.linkedin.com/learning/securing-the-use-of-generative-ai-in-your-organization",
	"https://www.linkedin.com/learning/openai-api-function-calling",
	"https://www.linkedin.com/learning/openai-api-fine-tuning-2023",
	"https://www.linkedin.com/learning/leveraging-ai-for-governance-risk-and-compliance",
	"https://www.linkedin.com/learning/introduction-to-ai-governance",
	"https://www.linkedin.com/learning/leveraging-ai-for-security-testing",
	"https://www.linkedin.com/learning/amplify-your-critical-thinking-with-generative-ai",
	"https://www.linkedin.com/learning/introduction-to-mlsecops",
	"https://www.linkedin.com/learning/openai-api-working-with-files",
	"https://www.linkedin.com/learning/openai-api-image-generation",
	"https://www.linkedin.com/learning/openai-api-vision",
	"https://www.linkedin.com/learning/openai-api-moderation-asi",
	"https://www.linkedin.com/learning/openai-api-introduction",
}

func buildCoursesFromLinks(courses *[]Course) {

	fmt.Println("Parsing courses")
	fmt.Println(len(li_ai_links))
	for _, n := range li_ai_links {
		data := Course{}
		data.link = n
		*courses = append(*courses, data)
	}
	fmt.Println("Building courses, done.")
	fmt.Println(len(*courses))
}

func parseDetailsFromLinks(ctx *context.Context, courses *[]Course) {

	for i, course := range *courses {
		fmt.Println(course.link)
		var ok bool
		data := Course{}
		time.Sleep(2 * time.Second)
		//time.Sleep(4 * time.Second)
		err := chromedp.Run(*ctx,
			chromedp.Navigate(course.link),
			// logged out
			chromedp.Text(`div.show-more-less-html__markup`, &data.details, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			// logged in?
			//chromedp.Text(`p.t-16`, &details, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			chromedp.Text(`h1.top-card-layout__title`, &data.title, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			chromedp.Text(`section > div > div > div > h2 > div:nth-child(1)`, &data.author, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			chromedp.Text(`section > div > div > div > h2 > div:nth-child(2) > span:nth-child(1)`, &data.duration, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			chromedp.Text(`section > div > div > div > h2 > div:nth-child(2) > span:nth-child(3)`, &data.releasedDate, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
			chromedp.AttributeValue(`img.top-card__image`, "src", &data.img, &ok, chromedp.NodeVisible, chromedp.ByQuery, chromedp.AtLeast(0)),
		)

		if err != nil {
			// ignore error
			fmt.Println(course.link)
			fmt.Println(err)
		}

		//DEBUG
		//time.Sleep(300 * time.Second)

		(*courses)[i].title = data.title
		(*courses)[i].author = data.author
		(*courses)[i].duration = data.duration
		(*courses)[i].releasedDate = data.releasedDate
		(*courses)[i].details = data.details
		(*courses)[i].img = data.img
	}

}

// HTML3 -
const HTML3 string = `
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
	<style>
			.body {
				margin: auto;
				width: 800px;
				font-family: Helvetica, Arial, sans-serif;
				line-height: 140%;
			}
			
			.introduction {
				max-width: 800px;
			}
			
			#top {
				text-align: center;
			}
			
			ul {
				list-style-type: none;
				margin-bottom: 15px;
				padding-inline-start: 0px;
			}
			
			li {
				list-style-type: none;
				margin-bottom: 5px;
			}
			
			img {
				width: 300px;
			}
			
			.completed {
				margin-bottom: 20px;
			}
			
			.details {
				max-width: 800px;
				clear: both;
			}
			
			.topbottom {
				margin-top: 20px;
				margin-bottom: 50px;
			}
			
			.leftside {
				float: left;
			}
			
			.rightside {
				float: right;
				padding-top: 20px;
			}
			
			.rightside li {
				text-align: right;
			}
			
			@media only screen and (max-width: 800px) {
				.body {
					margin: auto;
					width: 300px;
				} 
			
				ul {
					padding: 0;
				}
			
				.mainul {
					padding-bottom: 50px;
				}
			
				img {
					width: 300px;
				}
			
				.details {
					max-width: 300px;
				}
			
				.leftside {
					float: none;
				}
				
				.rightside {
					float: none;
					padding-top: 20px;
				}
				
				.rightside li {
					text-align: left;
				}
				
			}
		</style>
  </head>
  <body class="body">
    <main>
    <article class="page">
      <h1  id="top">LinkedIn Learning Courses Completed</h1>

      <div class="introduction">
      <p>
      This a summary of all the Linked-In courses offered by LinkedIn Learning for Keysight. 
      </p>
      <p>
      This list is generated from a tool called "main.go" that can be found
      <a
        href="https://github.com/alpiepho/pup-learning"
        target="_blank"
        rel="noreferrer"
      >here</a>.  This tool needs to be run manually to parse the LinkedIn Learning
      site to gather the list of courses from Keysight AI.
      </p>
      </div>
`

// HTML4 -
const HTML4 string = `
    <div id="bottom"></div>
    </article>
  </body>
</html>
`

func buildHTMLFromLinks(courses []Course, totalH int, totalM int, stages bool, stagehtml bool) {
	if !stages || stagehtml {
		today := time.Now().Local()
		var b strings.Builder
		fmt.Fprintf(&b, "%s", HTML3)
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
		fmt.Fprintf(&b, "%s", HTML4)

		err := writeToFile("./ai_list/index.html", b.String())
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
	fromlinks := flag.Bool("fromlinks", false, "a bool")
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

	// TODO LI-AI-LIST
	if *fromlinks {
		buildCoursesFromLinks(&courses)
		//saveThumbs(&ctx, &courses, *nopngs, *stages, *stagethumbs)
		if *getexfiles {
			parseDetailsFromLinks(&ctx, &courses)
			logoutAuto(&ctx)
		} else {
			// logout doesnt work (needs cookie) so try new browser
			//doLogout(&ctx, *stages, *stagelogout)
			cancel()
			ctx2, cancel2 := chromedp.NewExecAllocator(context.Background(), opts...)
			defer cancel2()
			ctx2, cancel2 = chromedp.NewContext(ctx2)
			defer cancel2()
			parseDetailsFromLinks(&ctx2, &courses)
		}
		if !*nosort {
			sort.Sort(ByCompleted(courses))
		}
		totalH, totalM := buildTimes(courses, *stages, *stagetimes)

		buildHTMLFromLinks(courses, totalH, totalM, *stages, *stagehtml)

	} else {
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
}
