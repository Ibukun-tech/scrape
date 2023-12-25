package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Post struct {
	Topic          string   `json:"topic"`
	MemoryVerse    string   `json:"memoryVerse"`
	BibleVerse     string   `json:"bibleVerse"`
	BibleVerseBody []string `json:"bibleVerseBody"`
	BibleInOneYear string   `json:"bibleInOneYear"`
	BodyMessage    []string `json:"bodyMessage"`
	Point          string   `json:"Point"`
	PointBody      string   `json:"pointBody"`
	HymnTitle      string   `json:"hymnTitle"`
	HymnBody       []string `json:"hymnBody"`
}

var (
	TextFilePost *string
	TextFileHtml *string
)

func GetCurrentDay() string {
	year, month, date := time.Now().Date()
	return fmt.Sprintf("%d-%s-%d", date, month, year)
}
func GetCurrentText() string {
	year, month, date := time.Now().Date()
	return fmt.Sprintf("%d-%s-%d.txt", date, month, year)
}
func init() {
	TextFilePost = flag.String("text", GetCurrentText(), "The text file where the post is stored")
	// TextFileHtml = flag.String("html", "post.html", "The Html file for the post")
}
func String(p *Post) string {
	return fmt.Sprintf("*%s*\n\n*MEMORY VERSE*\n%s\n\n*BIBLE READING*\n%s\n\n%s\n\n*MESSAGE*\n%s\n\n*%s*\n%s\n\n*HYMN*\n%s\n\n%s\n\n*BIBLE IN ONE YEAR*\n%s\n", p.Topic, p.MemoryVerse, p.BibleVerse, strings.Join(p.BibleVerseBody, "\n\n"), strings.Join(p.BodyMessage, "\n\n"), p.Point, p.PointBody, p.HymnTitle, strings.Join(p.HymnBody, "\n\n"), p.BibleInOneYear)
}
func (p *Post) SaveToText(st string) {
	file, err := os.Create(st)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	file.WriteString(String(p))
}
func main() {
	currentDate := GetCurrentDay()
	p := &Post{}

	date := flag.String("date", currentDate, "date to scrape")
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("did not connect", err)
	})
	c.OnResponse(func(r *colly.Response) {
		log.Println("response", r.Request.URL)
	})
	c.OnHTML(".et_pb_text_inner", func(h *colly.HTMLElement) {

		if !strings.Contains(h.Request.URL.String(), *date) {
			return
		}
		state := 0
		h.ForEach(".et_pb_text_inner h2, .et_pb_text_inner p:not(.has-text-align-center)", func(_ int, h *colly.HTMLElement) {
			// fmt.Println(h.Text)
			switch state {
			case 0:
				if h.Name == "h2" {
					state += 1
				}
			case 1:
				// fmt.Println(h.Text)
				p.Topic = h.Text
				state += 1
			case 2:
				// fmt.Println(h.Text)
				p.MemoryVerse = h.Text
				state += 1
			case 3:
				p.BibleVerse = h.Text
				state += 1
			case 4:
				if strings.Contains(h.Text, "YEAR") {
					p.BibleInOneYear = h.Text
					state += 1
				} else {
					p.BibleVerseBody = append(p.BibleVerseBody, h.Text)
				}
			case 5:
				if h.Name == "h2" {
					state += 1
				} else {
					state -= 1
				}
			case 6:
				if h.Name == "h2" {
					if strings.Contains(h.Text, "POINT") {
						p.Point = h.Text
						state += 1
					}
				} else {
					p.BodyMessage = append(p.BodyMessage, h.Text)
				}
			case 7:
				p.PointBody = h.Text
				state += 1
			case 8:
				if h.Name == "h2" {
					p.HymnTitle = h.Text
					state += 1
				} else {
					state = -1
				}

			case 9:
				html, _ := h.DOM.Html()
				st := strings.Replace(html, "<br/>", "\n", -1)
				if (len(st) > 0 && (st[0] >= '0' && st[0] <= '9')) || strings.Contains(st, "Refrain") || strings.Contains(st, "Chorus") {
					p.HymnBody = append(p.HymnBody, st)
				} else {
					return
				}
			default:
				fmt.Println("There is an error going on")
				return
			}
		})
	})
	URL := fmt.Sprintf("https://flatimes.com/open-heaven-%s/", currentDate)
	c.Visit(URL)
	c.Wait()
	fmt.Println(p)
	p.SaveToText(*TextFilePost)
	fmt.Println("created the file and written it there")
}
