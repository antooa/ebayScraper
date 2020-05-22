package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"log"
	"os"
	"strings"
)

type Product struct {
	Name  string            `json:"name"`
	Img   string            `json:"img"`
	Attrs map[string]string `json:"attrs"`
}

func main() {
	//url := os.Getenv("EBAY_CATEGORY_URL")
	url := "/b/PC-Laptops-Netbooks/177/bn_317584"
	c := colly.NewCollector(
		colly.Async(true),
	)

	extensions.RandomUserAgent(c)

	products := make([]Product, 0)

	c.OnHTML("ul.b-list__items_nofooter", func(element *colly.HTMLElement) {
		element.ForEach("li.s-item", func(_ int, elem *colly.HTMLElement) {
			p := Product{Attrs: make(map[string]string)}
			p.Name = elem.ChildText("h3.s-item__title")

			img1 := elem.ChildAttr("img.s-item__image-img", "src")
			img2 := elem.ChildAttr("img.s-item__image-img", "data-src")

			if strings.Contains(img1, "ir.ebaystatic.com") {
				p.Img = img2
			} else {
				p.Img = img1
			}

			elem.ForEach("span.s-item__detail.s-item__detail--secondary", func(_ int, e *colly.HTMLElement) {
				attr := e.ChildText("span.s-item__dynamic")
				if attr == "" {
					return
				}

				kv := strings.SplitN(attr, ": ", 2)
				key := kv[0]
				val := kv[1]
				p.Attrs[key] = val

			})
			products = append(products, p)
		})
	})

	c.OnResponse(func(r *colly.Response) {

		log.Println("Visiting ", r.Request.URL.String())
		doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(r.Body))
		if err != nil {
			log.Fatal(err)
		}

		el := doc.Find("ul.b-list__items_nofooter")
		if len(el.Nodes) == 0 {
			log.Println("No items found on ", r.Request.URL.String())
		}

	})

	for i := 1; i <= 220; i++ {
		url := fmt.Sprintf("https://www.ebay.com%s?_pgn=%d", url, i)
		err := c.Visit(url)
		if err != nil {
			log.Fatal(err)
			break
		}
	}

	c.Wait()

	file, err := os.OpenFile("result.json", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(products)
	if err != nil {
		log.Fatal(err)
	}
}

