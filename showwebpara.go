package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func showPara(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()       //解析参数，默认是不会解析的
	fmt.Println(r)      //打印请求信息
	fmt.Println(r.Form) //打印所有参数信息
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)

	for k, v := range r.Form { //逐个打印各参数信息
		fmt.Println("key:", k)
		fmt.Println("val:",v)
	}
}

func main() {
	http.HandleFunc("/", showPara)           //设置访问的路由
	err := http.ListenAndServe(":6060", nil) //设置监听的端口
	if err != nil {
		log.Fatal("服务启动失败: ", err)
	}
}
