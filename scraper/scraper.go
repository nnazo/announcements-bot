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
}

func (ptr *Scraper) Setup() {
	ptr.URL = "https://natalie.mu/comic/tag/43"
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
				fmt.Println(article)
				if ptr.newest.URL == "" {
					ptr.newest = article
					return false
				}
				if ptr.newest == article {
					return false
				}
				ptr.articles = append(ptr.articles, article)
				return true
			})
		})
	})
}

func (ptr *Scraper) Newest() Article {
	return ptr.newest
}

func (ptr *Scraper) FetchNewArticles() []Article {
	ptr.articles = make([]Article, 0)

	ptr.c.Visit(ptr.URL)
	ptr.c.Wait()

	if len(ptr.articles) > 0 {
		ptr.newest = ptr.articles[0]
	}

	return ptr.articles
}

func getArticle(e *colly.HTMLElement) Article {
	return Article{
		URL:      e.ChildAttr("a", "href"),
		Title:    e.ChildText(".NA_title"),
		Summary:  e.ChildText(".NA_summary"),
		Date:     e.ChildText(".NA_date"),
		Comments: e.ChildText(".NA_comments"),
	}
}
