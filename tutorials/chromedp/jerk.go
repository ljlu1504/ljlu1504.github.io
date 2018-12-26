package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp/kb"
	"github.com/chromedp/chromedp/runner"
	"io/ioutil"
	"log"
	"time"
	
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {
	var err error
	
	// create context
	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// create chrome instance
	runnerOps := chromedp.WithRunnerOptions(
		runner.Path(`C:\Users\zhouqing1\AppData\Local\Google\Chrome\Application\chrome.exe`),
		runner.Flag("no-default-browser-check", true),
		runner.Flag("no-sandbox", true),
		//runner.WindowSize(750, 1624),
		//runner.UserAgent(`Mozilla/5.0 (iPhone; CPU OS 11_0 like Mac OS X) AppleWebKit/604.1.25 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1`),
	)
	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf), runnerOps)
	if err != nil {
		log.Fatal(err)
	}
	var nodes []*cdp.Node
	err = c.Run(ctxt, tranBaiduForMojotv(&nodes))
	if err != nil {
		log.Fatal(err)
	}
	for _,v := range nodes {
		nodeV := v.NodeValue
		fmt.Println(nodeV)
	}
	
	// shutdown chrome
	err = c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}
	
	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func visitMojoTvDotCn(url string, elementHref, pageTitle, iFrameHtml *string) chromedp.Tasks {
	//临时放图片buf
	var buf []byte
	return chromedp.Tasks{
		//跳转到页面
		chromedp.Navigate(url),
		//chromedp.Sleep(2 * time.Second),
		//等待博客正文显示
		chromedp.WaitVisible(`#post`, chromedp.ByQuery),
		//滑动页面到google adsense 广告
		chromedp.ScrollIntoView(`ins`, chromedp.ByQuery),
		chromedp.Screenshot(`#post`, &buf, chromedp.ByQuery, chromedp.NodeVisible),
		//截图到文件
		chromedp.Sleep(2 * time.Second),
		chromedp.ActionFunc(func(context.Context, cdp.Executor) error {
			return ioutil.WriteFile("mojotv_local.png", buf, 0644)
		}),
		chromedp.ScrollIntoView(`#copyright`, chromedp.ByID),
		//等待mojotv google广告展示出来
		chromedp.WaitVisible(`#post__title`, chromedp.ByID),
		chromedp.Sleep(2 * time.Second),
		
		//获取我的google adsense 广告代码
		chromedp.InnerHTML(`#post__title`, iFrameHtml, chromedp.ByID),
		//跳转到我的bilibili网站
		chromedp.Sleep(5 * time.Second),
		
		chromedp.Click("#copyright > a:nth-child(3)", chromedp.NodeVisible),
		//等待则个页面显现出来
		chromedp.WaitVisible(`#page`, chromedp.ByQuery),
		//在chrome浏览器页面里执行javascript
		chromedp.Evaluate(`document.title`, pageTitle),
		chromedp.Screenshot(`#page`, &buf, chromedp.ByQuery, chromedp.NodeVisible),
		chromedp.Sleep(5 * time.Second),
		
		//截取bili网页图片
		chromedp.ActionFunc(func(context.Context, cdp.Executor) error {
			return ioutil.WriteFile("bili_local.png", buf, 0644)
		}),
		//获取bilibili网页的标题
		chromedp.JavascriptAttribute(`a`, "href", elementHref, chromedp.ByQuery),
	}
}
func tranBaiduForMojotv(nodes *[]*cdp.Node) chromedp.Tasks {
	//临时放图片buf
	url:= `https://www.baidu.com`
	slll := `//h3/a[contains(text(),'一个热爱分享的')]`
	netxbtn := `//a[@class='n']`
	return chromedp.Tasks{
		//跳转到页面
		chromedp.Navigate(url),
		//等待博客正文显示
		chromedp.WaitVisible(`#kw`, chromedp.ByID),
		chromedp.SendKeys(`#kw`,"mojotv",chromedp.ByID),
		chromedp.KeyAction(kb.Enter),
		
		chromedp.WaitVisible("h3",chromedp.ByQueryAll),
		chromedp.Nodes(slll,nodes),
		chromedp.Click(slll),
		
		chromedp.Sleep(time.Second*5),
		
		
	}
}
