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
}

type Scraper struct {
	URL      string
	c        *colly.Collector
	articles []Article
	newest   Article
	stop     Article
}

var (
	articles = make([]Article, 0)
)

func (ptr *Scraper) Setup(url string) {
	ptr.articles = make([]Article, 0)
	ptr.URL = url
	ptr.c = colly.NewCollector(
		colly.AllowedDomains("natalie.mu"),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)
	ptr.c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	ptr.c.OnHTML("#NA_main", func(e *colly.HTMLElement) {
		e.ForEach(".NA_articleList", func(i int, e *colly.HTMLElement) {
			e.ForEachWithBreak("li", func(j int, e *colly.HTMLElement) bool {
				article := getArticle(e)
				fmt.Println("checking", article.URL)
				if ptr.stop.URL == "" {
					ptr.stop = ptr.newest
				}
				fmt.Println("\tstop", ptr.stop.URL)
				fmt.Println("\tnewest", ptr.newest.URL)

				if ptr.stop.URL != "" {
					if ptr.stop.URL != article.URL {
						articles = append(articles, article)
						fmt.Println("new article found", articles)
					} else {
						fmt.Println("\tsame as newest")
						ptr.stop = ptr.newest
						return false
					}

					if j == 0 {
						ptr.newest = article
					}

					return true
				} else {
					fmt.Println("\tinitialize newest")
					ptr.newest = article
					ptr.stop = article
				}

				return false
			})
		})
	})
}

func (ptr *Scraper) FetchNewArticles() []Article {
	// i have to copy from a different slice because for
	// some reason the OnHTML callback gets a different address
	// for ptr.articles and doesn't reflect changes here
	articles = articles[:0]

	if ptr.c != nil {
		ptr.c.Visit(ptr.URL)
		ptr.c.Wait()
	} else {
		panic(fmt.Errorf("nil collector"))
	}

	ptr.articles = articles

	return ptr.articles
}

func getArticle(e *colly.HTMLElement) Article {
	return Article{
		URL:      e.ChildAttr("a", "href"),
		Title:    e.ChildText(".NA_title"),
		Summary:  e.ChildText(".NA_summary"),
		Date:     e.ChildText(".NA_date"),
		Comments: e.ChildText(".NA_comment2"),
	}
}
