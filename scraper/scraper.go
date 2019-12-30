package scraper

import (
	"fmt"

	"github.com/gocolly/colly"
)

type Article struct {
	URL      string
	Title    string
	Summary  string
	Date     string
	Comments string
	Sent     bool
}

type Scraper struct {
	URL      string `json:"url"`
	c        *colly.Collector
	Articles []*Article
}

func (ptr *Scraper) Setup() {
	ptr.Articles = make([]*Article, 0)
	ptr.c = colly.NewCollector(
		colly.AllowedDomains("natalie.mu"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)
	ptr.c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	ptr.c.OnHTML("#NA_main", func(e *colly.HTMLElement) {
		e.ForEach(".NA_articleList", func(liNdx int, e *colly.HTMLElement) {
			if liNdx > 0 {
				return
			}
			articles := make([]*Article, 0)
			oldHTML := false
			e.ForEach("li", func(j int, e *colly.HTMLElement) {
				article := getArticle(e)
				articles = append(articles, article)
			})
			if len(ptr.Articles) < 1 {
				fmt.Println("initializing articles")
				for i := 0; i < len(articles); i++ {
					articles[i].Sent = true
				}
			} else {
				for i := range articles {
					ndx := -1
					for _, a := range ptr.Articles {
						if a.URL == articles[i].URL {
							articles[i].Sent = a.Sent
							ndx = i
							break
						}
					}
					if ndx < 0 {
						if i < (len(articles) / 2) {
							fmt.Println("\tfound unsent article", articles[i].URL)
						} else {
							oldHTML = true
						}
					}
				}
			}

			if !oldHTML {
				ptr.Articles = articles
			}
		})
	})
}

func (ptr *Scraper) UpdateArticles() {
	if ptr.c != nil {
		ptr.c.Visit(ptr.URL)
		ptr.c.Wait()
	} else {
		panic("nil collector")
	}
}

func getArticle(e *colly.HTMLElement) *Article {
	return &Article{
		URL:      e.ChildAttr("a", "href"),
		Title:    e.ChildText(".NA_title"),
		Summary:  e.ChildText(".NA_summary"),
		Date:     e.ChildText(".NA_date"),
		Comments: e.ChildText(".NA_comment2"),
	}
}
