package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

/**
多看视频爬虫
*/
type ResolutionType int8

const (
	BQ ResolutionType = iota
	GC
	CQ
	LG
)

func main() {
	err := downloadFile("/Users/download/", "https://haokan.baidu.com/v?vid=14514233269277091964&tab=recommend", BQ)
	if err != nil {
		log.Fatal("下载失败：", err)
	}
	log.Println("下载成功")
}

/**

 */
func downloadFile(filepath string, url string, resolutionType ResolutionType) (err error) {

	_, videoUrl, err := getVideoUrl(url, resolutionType)

	if err != nil {
		log.Fatal("解析视频地址异常：", err)
	}
	split := strings.Split(strings.Split(videoUrl, "?")[0], "/")
	fileName := split[len(split)-1]
	filepath = filepath + "/" + fileName
	// Create the file

	_, err = os.Stat(filepath)

	if err != nil {
		_, err := os.Create(filepath)
		if err != nil {
			log.Fatal("创建文件失败：", err)
		}
	}
	out, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	defer out.Close()

	fmt.Println(videoUrl)
	// Get the data
	resp, err := http.Get(videoUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

/**
1、根据网页url获取页面元素
2、解析页面元素，获取视频地址
3、根据视频地址进行下载
*/
func getVideoUrl(url string, resolutionType ResolutionType) (string, string, error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("初始化请求失败：", err)
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")

	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("请求页面出错:", err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("返回码：", resp.StatusCode)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("读取页面内容出错：", err)
	}
	result := string(bytes)
	result = strings.Replace(result, "\\/", "/", -1)

	nameRgx := regexp.MustCompile(`<meta itemprop="name" name="title" content="(.*?)">`)
	bqRgx := regexp.MustCompile(`"title":"\\u6807\\u6e05","url":"(.*?)",`)
	gqRgx := regexp.MustCompile(`"title":"\\u9ad8\\u6e05","url":"(.*?)",`)
	cqRgx := regexp.MustCompile(`"title":"\\u8d85\\u6e05","url":"(.*?)",`)
	lgRgx := regexp.MustCompile(`"title":"\\u84dd\\u5149","url":"(.*?)",`)

	titles := nameRgx.FindStringSubmatch(result)

	var videoTitle string

	if len(titles) > 0 {
		videoTitle = titles[1]
	} else {
		videoTitle = ""
	}

	var videoUrl string
	var newError error

	switch resolutionType {
	case BQ:
		bq := bqRgx.FindStringSubmatch(result)
		videoUrl = bq[1]
	case GC:
		gq := gqRgx.FindStringSubmatch(result)
		videoUrl = gq[1]
	case CQ:
		cq := cqRgx.FindStringSubmatch(result)
		videoUrl = cq[1]
	case LG:
		lg := lgRgx.FindStringSubmatch(result)
		videoUrl = lg[1]
	default:
		newError = errors.New("视频分辨率类型输入有误")
	}
	return videoTitle, videoUrl, newError

}
