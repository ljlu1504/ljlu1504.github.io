package util

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"time"
)

func SpiderRedditProgramming() error {
	doc, err := downloadHtml("https://www.reddit.com/r/programming/")
	if err != nil {
		return err
	}
	fmt.Println(doc.Contents().Text())
	pipe := RedisClient.Pipeline()
	// Find the review items
	skey := time.Now().Format("redditnews-2006-01-02")
	hkey := "redditnews"
	doc.Find("#siteTable > div.link > div.entry.unvoted > div.top-matter > p.title > a").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		fmt.Print(url)
		pipe.SAdd(skey, url)
		if RedisClient.HGet(hkey, url).Val() == "" {
			titleEn := s.Text()
			titleZh := TranslateEn2Ch(titleEn)
			timeString := time.Now().Format("2006-01-02")
			newsItem := NewsItem{titleZh, titleEn, url, timeString}
			if bytes, err := json.Marshal(newsItem); err == nil {
				pipe.HSet(hkey, url, bytes)
			}
			time.Sleep(time.Microsecond * 100)
		}
	})
	pipe.Expire(skey, time.Hour*12)
	pipe.Exec()
	return nil
}
