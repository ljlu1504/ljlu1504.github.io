package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dejavuzhou/dejavuzhou.github.io/config"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

const hackNewsUrl = "https://news.ycombinator.com/news"

type NewsItem struct {
	TitleZh string `json:"titleZh"`
	TitleEn string `json:"titleEn"`
	Url     string `json:"url"`
	Date    string `json:"date"`
}

func downloadHtml(url string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("cookie", config.HTTP_COOKIE)
	req.Header.Set("User-Agent", config.HTTP_USER_AGENT)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("the get request's response code is not 200")
	}
	defer res.Body.Close()
	return goquery.NewDocumentFromReader(res.Body)
}
func SpiderHackNews() error {
	//stories := []item{}
	// Instantiate default collector
	doc, err := downloadHtml(hackNewsUrl)
	if err != nil {
		return err
	}
	pipe := RedisClient.Pipeline()
	// Find the review items
	skey := time.Now().Format("hacknews:2006-01-02")
	hkey := time.Now().Format("hacknews:2006-01")
	doc.Find("a.storylink").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
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
	pipe.Expire(skey, time.Hour*24)
	pipe.Expire(hkey, time.Hour*24)
	pipe.Exec()
	return nil
}
func fetchRedisDataHackNews(itemType string) ([]NewsItem, error) {
	skey := time.Now().Format(itemType+":2006-01-02")
	urls, err := RedisClient.SMembers(skey).Result()
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, errors.New("爬虫没有内容")
	}
	hkey := time.Now().Format(itemType+":2006-01")
	
	jsonStrings, err := RedisClient.HMGet(hkey, urls...).Result()
	if err != nil {
		return nil, err
	}
	newsItems := []NewsItem{}
	for _, item := range jsonStrings {
		if string, ok := item.(string); ok {
			items := NewsItem{}
			json.Unmarshal([]byte(string), &items)
			newsItems = append(newsItems, items)
		}
	}
	
	return newsItems, err
}
func ParsemarkdownHacknews() error {
	
	newsItems, err := fetchRedisDataHackNews("hacknews")
	if err != nil {
		return err
	}
	showItems, err := fetchRedisDataHackNews("hackshows")
	if err != nil {
		return err
	}
	if len(newsItems) < 1 {
		return nil
	}
	
	tmpl, err := template.ParseFiles("util/hacknews.tpl") //解析模板文件
	day := time.Now().Format("2006-01-02")
	mdFile := fmt.Sprintf("_posts/hacknews/%s-hacknews.md", day)
	
	file, err := os.Create(mdFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	data := struct {
		Day  string
		News []NewsItem
		Shows []NewsItem
	}{day, newsItems, showItems}
	err = tmpl.Execute(file, data) //执行模板的merger操作
	return err
}



func SpiderHackShows() error {
	//stories := []item{}
	// Instantiate default collector
	doc, err := downloadHtml("https://news.ycombinator.com/show")
	if err != nil {
		return err
	}
	pipe := RedisClient.Pipeline()
	// Find the review items
	skey := time.Now().Format("hackshows:2006-01-02")
	hkey := time.Now().Format("hackshows:2006-01")
	doc.Find("a.storylink").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		if !strings.Contains(url,"http"){
			url = "https://news.ycombinator.com/" + url
		}
		pipe.SAdd(skey, url)
		if RedisClient.HGet(hkey, url).Val() == "" {
			titleEn := s.Text()
			titleEn = strings.Replace( titleEn,"Show HN:", "", -1)
			titleZh := TranslateEn2Ch(titleEn)
			timeString := time.Now().Format("2006-01-02")
			newsItem := NewsItem{titleZh, titleEn, url, timeString}
			if bytes, err := json.Marshal(newsItem); err == nil {
				pipe.HSet(hkey, url, bytes)
			}
			time.Sleep(time.Microsecond * 100)
		}
	})
	pipe.Expire(skey, time.Hour*24)
	pipe.Expire(hkey, time.Hour*24)
	pipe.Exec()
	return nil
}
