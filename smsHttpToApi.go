package main

/*
#cgo CFLAGS:  -I./
#cgo LDFLAGS: -L${SRCDIR} -l ImApi
#include <ImApi.h>
#include <stdio.h>
*/
import "C"

import (
	"fmt"
	"log"
	"net/http"
)

func httpToApi(w http.ResponseWriter, r *http.Request) {

	var to = ""
	var content = ""

	r.ParseForm()
	for k, v := range r.Form {
		fmt.Println(k, v)
		switch k {

		case "to":
			to = v[0]
			fmt.Println(k, v)
		case "content":
			content = v[0]
			fmt.Println(k, v)
		}
	}
	fmt.Println("------------------------")

	ret := int32(C.sendSM(C.CString(to), C.CString(content), C.long(5)))
	if ret != 0 {
		fmt.Fprintf(w, "send sms failed!")
	} else {
		fmt.Fprintf(w, "success")
	}

}

func main() {
	ret := int32(C.initWithDB(C.CString("xxx"), C.CString("xxx"), C.CString("xxx"), C.CString("xxx"), C.CString("xxx")))
	if ret != 0 {
		fmt.Println("connect to the sms-gateway error!")
		return
	} else {
		fmt.Println("sms-gateway connected!")
	}
	http.HandleFunc("/sms", httpToApi)       //设置访问的路由
	err := http.ListenAndServe(":8030", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
