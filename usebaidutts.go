package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

var (
	tokenUrl string = "https://openapi.baidu.com/oauth/2.0/token"
	ttsUrl   string = "http://tsn.baidu.com/text2audio"
	//tokenUrl string = "http://127.0.0.1:6060"
	//ttsUrl    string = "http://127.0.0.1:6060"

	apikey    string = "xxxxxxx"
	SecretKey string = "yyyyyyyyyyyyyyyyy"
	tex       string = "百度tts语音测试"
	tok       string = ""
	cuid      string = "tts.abc.com/test"
	ctp       string = "1"
	lan       string = "zh"
	spd       string = "5"
	pit       string = "5"
	vol       string = "5"
	per       string = "1"
	aue       string = "3"
)

func main() {
	v := url.Values{}
	v.Add("apikey", apikey)
	v.Add("SecretKey", SecretKey)
	v.Add("tex", url.QueryEscape(tex))
	tok, err := getToken()
	if err == nil {
		v.Add("tok", tok)
	} else {
		fmt.Println("get token fail", err)
		return
	}
	v.Add("cuid", cuid)
	v.Add("ctp", ctp)
	v.Add("lan", lan)
	v.Add("spd", spd)
	v.Add("pit", pit)
	v.Add("vol", vol)
	v.Add("per", per)
	v.Add("aue", aue)

	res, err := http.PostForm(ttsUrl, v)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	resheader := fmt.Sprintln(res.Header)
	if strings.Contains(resheader, "json") {
		fmt.Println(res.Body)
	}
	defer res.Body.Close()
	fmt.Println("post send success")
	f, err := os.Create("tt.mp3")
	if err != nil {
		panic(err)
	}
	io.Copy(f, res.Body)
}

func getToken() (string, error) {
	pageBuf := make([]byte, 4096)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	//生成请求
	para := url.Values{}
	para.Add("grant_type", "client_credentials")
	para.Add("client_id", apikey)
	para.Add("client_secret", SecretKey)
	spara := para.Encode()
	reqest, err := http.NewRequest("POST", tokenUrl, strings.NewReader(spara))

	reqest.Header.Add("Pragma", "no-cache")
	reqest.Header.Add("Cache-Control", "no-cache")
	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		fmt.Println("生成网页请求失败", tokenUrl, err)
		return "", err
	}

	//处理返回结果
	response, err := client.Do(reqest)

	//TODO 需要等待数据返回吗？选等一秒吧
	time.Sleep(time.Second * 1)

	if err != nil {
		fmt.Println("获取网页内容失败", tokenUrl, err)
		return "", err
	}

	defer response.Body.Close()
	//TODO 处理网页返回的状态码
	if response.StatusCode != 200 && response.StatusCode != 302 {
		return "", errors.New("读网页错误，错误码：" + strconv.Itoa(response.StatusCode))
	}

	var n int
	var page string
	for err = nil; err == nil; {
		n, err = response.Body.Read(pageBuf)
		page = page + string(pageBuf[:n])
	}
	// fmt.Println(page)
	//TODO解析返回的json串
	_, err = jsonparser.GetString([]byte(page), "error")
	if err == nil {
		fmt.Println("获取token失败", err)
		return "fail", err
	}

	token, err := jsonparser.GetString([]byte(page), "access_token")
	iftts, err2 := jsonparser.GetString([]byte(page), "scope")
	if err == nil && err2 == nil && strings.Contains(iftts, "audio_tts_post") {
		return token, nil
	} else {
		fmt.Println("获取token失败", err)
		return "", errors.New("get token fail")
	}
}
