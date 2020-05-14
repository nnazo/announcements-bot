package scraper

import (
	"github.com/gocolly/colly"
)

type Article struct {
	URL      string
	Image    string
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
	// ptr.c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String())
	// })

	ptr.c.OnHTML("main", func(e *colly.HTMLElement) {
		articles := make([]*Article, 0)
		oldHTML := false

		e.ForEach(".NA_section-list", func(_ int, e *colly.HTMLElement) {
			e.ForEach(".NA_card_wrapper", func(_ int, e *colly.HTMLElement) {
				e.ForEach(".NA_card-l", func(_ int, e *colly.HTMLElement) {
					article := getArticle(e)
					articles = append(articles, article)
				})
			})
		})

		if len(ptr.Articles) < 1 {
			// fmt.Println("initializing articles")
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
					if i >= (len(articles) / 2) {
						oldHTML = true
					} /* else {
						fmt.Println("\tfound unsent article", articles[i].URL)
					}*/
				}
			}
		}

		if !oldHTML {
			ptr.Articles = articles
		}
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
	urls := e.ChildAttrs("a", "href")
	imgs := e.ChildAttrs("img", "data-src")
	return &Article{
		URL:     urls[len(urls)-1],
		Image:   imgs[len(imgs)-1],
		Title:   e.ChildText(".NA_card_title"),
		Summary: e.ChildText(".NA_card_summary"),
		Date:    e.ChildText(".NA_card_date"),
		//Comments: e.ChildText(".NA_comment2"),
	}
}
