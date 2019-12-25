package scraper

import (
	"fmt"
	"strconv"
	"strings"

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
}

var (
	articles []Article = make([]Article, 0)
)

func (ptr *Scraper) Setup(url string) {
	ptr.articles = make([]Article, 0)
	ptr.URL = url
	ptr.c = colly.NewCollector(
		colly.AllowedDomains("natalie.mu"),
		colly.DisallowedDomains("store.natalie.mu"),
		colly.AllowURLRevisit(),
		colly.Async(false),
	)
	ptr.c.Limit(&colly.LimitRule{
		Parallelism: 2,
	})
	ptr.c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	ptr.c.OnHTML("#NA_main", func(e *colly.HTMLElement) {
		e.ForEach(".NA_articleList", func(i int, e *colly.HTMLElement) {
			e.ForEachWithBreak("li", func(j int, e *colly.HTMLElement) bool {
				article := getArticle(e)
				fmt.Println("checking", article.URL)
				if ptr.newest.URL != "" {
					s1 := strings.Split(article.URL, "/")
					id1, _ := strconv.Atoi(s1[len(s1)-1])
					s2 := strings.Split(ptr.newest.URL, "/")
					id2, _ := strconv.Atoi(s2[len(s1)-1])

					if id1 < id2 {
						return false
					}
				}
				if ptr.newest.URL != "" && ptr.newest.URL != article.URL {
					articles = append(articles, article)
					fmt.Println("new article found", ptr.articles)
					ptr.newest = article
					return true
				}
				if ptr.newest.URL == "" {
					fmt.Println("\tinitialize newest")
					ptr.newest = article
				} else {
					fmt.Println("\tsame as newest")
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

	ptr.c.Visit(ptr.URL)
	ptr.c.Wait()

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
