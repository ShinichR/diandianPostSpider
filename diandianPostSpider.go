package main

import (
	"fmt"
	"github.com/hu17889/go_spider/core/common/page"
	"github.com/hu17889/go_spider/core/pipeline"
	"github.com/hu17889/go_spider/core/spider"
	"html"
	"os"
	"regexp"
	"strings"
	//"sync"
)

type MyPageProcesser struct {
	visit_url map[string]int
	newurl    chan string
}

func NewMyPageProcesser() *MyPageProcesser {
	return &MyPageProcesser{}
}

func (this *MyPageProcesser) Init() {
	this.visit_url = make(map[string]int)
	//this.newurl = make(chan string, 10)
	//fmt.Println("process init")
}

// Parse html dom here and record the parse result that we want to Page.
// Package goquery (http://godoc.org/github.com/PuerkitoBio/goquery) is used to parse html.
func (this *MyPageProcesser) Process(p *page.Page) {
	if !p.IsSucc() {
		println(p.Errormsg())
		return
	}
	var crawok bool
	crawok = false
	//query := p.GetHtmlParser()
	//var urls []string

	//fmt.Println(p.GetBodyStr())
	re := regexp.MustCompile(`<a href="(.*?)">(.*?)`)

	sectUrlsTemp := re.FindAllSubmatch([]byte(p.GetBodyStr()), -1)

	for _, url := range sectUrlsTemp {
		for _, url1 := range url {
			crawok = true

			http_index := strings.Index(string(url1), "http://shinichr.diandian.com")

			http_note := strings.Index(string(url1), "\"")
			http_quote := strings.Index(string(url1), "#")

			if http_index >= 0 && http_quote < 0 {
				if http_note > 0 && http_note < http_index {
					continue
				}

				var http_url string
				if http_note <= 0 {
					http_url = string(url1)[http_index:]
				} else {
					http_url = string(url1)[http_index:http_note]
				}

				if this.visit_url[http_url] == 0 {
					this.visit_url[http_url] = 1

					fmt.Println("####unvisited:", http_url)
					//fmt.Println("###AddTargetRequest:", http_url)
					p.AddTargetRequest(http_url, "html")
				}
			}
		}
	}
	if crawok == false {
		fmt.Println("crawl false:*****************", p.GetRequest().GetUrl())
		http_page := strings.Index(p.GetRequest().GetUrl(), "http://shinichr.diandian.com/page")
		http_post := strings.Index(p.GetRequest().GetUrl(), "http://shinichr.diandian.com/post")
		fmt.Println("http_page:", http_page, "http_post:", http_post)
		if http_page >= 0 || http_post >= 0 {
			//this.visit_url[p.GetRequest().GetUrl()] = 0
			p.AddTargetRequest(p.GetRequest().GetUrl(), "html")
		}
	}
	http_index := strings.Index(p.GetRequest().GetUrl(), "http://shinichr.diandian.com/post/")

	//	rex, _ := regexp.Compile("\\/")
	//replaceurl := rex.ReplaceAllString(p.GetRequest().GetUrl(), ".")
	//fmt.Println("http_index=", http_index)
	//fmt.Println("replaceurl=", p.GetRequest().GetUrl()[http_index:], "....", http_index)
	if http_index >= 0 {

		cuturl := p.GetRequest().GetUrl()[34:]
		//fmt.Println("replaceurl=", cuturl)
		rex, _ := regexp.Compile("\\/")
		replaceurl := rex.ReplaceAllString(cuturl, ".")

		filedir := fmt.Sprintf("/home/shinichr/diandian_post/%s", replaceurl)

		fout, err := os.Create(filedir)
		if err != nil {
			fmt.Println(filedir, err)
			return
		}
		defer fout.Close()

		src := p.GetBodyStr()
		re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
		src = re.ReplaceAllStringFunc(src, strings.ToLower)

		//去除STYLE
		re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
		src = re.ReplaceAllString(src, "")

		//去除SCRIPT
		re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
		src = re.ReplaceAllString(src, "")

		//去除所有尖括号内的HTML代码，并换成换行符
		re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
		src = re.ReplaceAllString(src, "\n")

		//去除连续的换行符
		re, _ = regexp.Compile("\\s{2,}")
		src = re.ReplaceAllString(src, "\n")

		//fmt.Println(strings.TrimSpace(src))

		fout.WriteString(html.UnescapeString(src))
		fmt.Println("save file ", filedir)
	}
	//query.Find(`div[class="rich-content] div[class="post"] div[class="post-top"] div[class="post-content post-text"] a`).Each(func(i int, s *goquery.Selection) {
	/*query.Find("div.content").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		http_index := strings.Index(href, "http")
		http_url := href[http_index:]

		//fmt.Println("###url:\n", http_url, "=", this.visit_url[http_url])
		//this.newurl <- http_url
		if this.visit_url[http_url] == 0 {
			this.visit_url[http_url] = 1
			fmt.Println("###AddTargetRequest:", http_url)
			p.AddTargetRequest(http_url, "html")
		}
		urls = append(urls, href)
	})*/
	// these urls will be saved and crawed by other coroutines.
	/*doc, _ := goquery.NewDocument("http://shinichr.diandian.com")

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		fmt.Println("####href=", href)
	})*/
	//p.AddField("readme", readme)
}

func (this *MyPageProcesser) Finish() {
	fmt.Printf("TODO:before end spider \r\n")
}

func main() {

	pageProcess := NewMyPageProcesser()
	pageProcess.Init()
	diandianSpider := spider.NewSpider(pageProcess, "TaskName").
		AddUrl("http://shinichr.diandian.com/", "html"). // Start url, html is the responce type ("html" or "json" or "jsonp" or "text")
		AddPipeline(pipeline.NewPipelineConsole()).      // Print result on screen
		AddPipeline(pipeline.NewPipelineFile("/home/shinichr/spider.tmp")).
		SetThreadnum(10000) // Crawl request by three Coroutines

	for i := 2; i < 10; i++ {
		url := fmt.Sprintf("http://shinichr.diandian.com/page/%d", i)
		diandianSpider.AddUrl(url, "html")
	}

	diandianSpider.Run()

	for url, _ := range pageProcess.visit_url {
		fmt.Println("spider:", url)
	}

}
