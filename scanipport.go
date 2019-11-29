// scanipport.go
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//本程序用于扫描多个IP的所有端口是否有开放，IP地址列表存放在同一目录下的文本文件里，默认文件名为ips.txt
//扫描的方式是去尝试进行tcp连接，如果连接成功证明端口是开放的
//这种扫描方式很容易被防火墙认为是攻击从而阻断扫描机器的IP，因此主要用于对自家系统进行安全摸底扫描，扫描前把扫描机加入防火墙白名单
//如果扫描结果要存放到文件里，建议使用输出重定向到csv文件里，为了方便csv分栏，IP和端口的分隔符已替换为逗号
//协程数是针对每个IP设置的,不宜设置太大,设置太大会造成创建连接失败漏掉端口，设置的时候协程数*IP数不能大于65534，考虑其它程序有使用一部分连接，总数不要大于30000.

var connTimeOut = flag.Int("timeOut", 800, "连接超时时长毫秒")
var fileName = flag.String("file", "ips.txt", "存放要扫描的ip列表文件名,每个ip一行，不留空行")
var help = flag.String("h", "", "help")
var threads = flag.Int("ipThreads", 100, "扫描每个ip的线程数,总ip数剩以ipThreads不应该超过空余端口数")
var startPort = flag.Int("startPort", 1, "开始端口")
var endPort = flag.Int("endPort", 65535, "结束端口")

func main() {
	fmt.Println("简单的端口扫描工具，可同时对一批IP进行端口扫描，自由软件，随便用。")
	fmt.Println("")

	flag.Parse()
	FileRead(*fileName)
	time.Sleep(time.Second * 3)
	for runtime.NumGoroutine() > 1 {
		time.Sleep(time.Second * 3)
	}

}
func FileRead(name string) {

	fileObj, err := os.Open(name)
	if err != nil {
		fmt.Println("打开IP列表文件失败，请检查文件名是否正确，默认文件名为ips.txt")
		return
	}
	defer fileObj.Close()
	//在定义空的byte列表时尽量大一些，否则这种方式读取内容可能造成文件读取不完整
	buf := make([]byte, 4096)

	n, err := fileObj.Read(buf)
	if err != nil {
		fmt.Println("读取ip列表失败！")
		return
	}

	result := string(buf[:n])
	//fmt.Println("原始文件内容:", result)

	//对ip列表文件的原始内容进行处理，去除多余空格 ，统一换行符
	result = strings.TrimSpace(result)
	result = strings.Replace(result, "\r\n", "\n", -1)

	ips := strings.Split(result, "\n")
	fmt.Println("ip地址列表:", ips)
	scan(ips)

}

func scan(ips []string) {
	time.Sleep(time.Second)
	for _, ip := range ips {
		go scanAIP(ip)
	}
}

func scanAIP(ip string) {

	//通讯用缓冲管道500够用了吧。
	chIPPort := make(chan string, 2000)
	var port uint16

	for i := *threads; i > 0; i-- {
		go scanAPort(chIPPort)
	}

	for port = uint16(*startPort); port < uint16(*endPort); port++ {

		ipAndPort := ip + ":" + strconv.FormatUint(uint64(port), 10)
		chIPPort <- ipAndPort

	}
	close(chIPPort)

}

func scanAPort(chStrIPPort chan string) {

	var conn net.Conn
	var tcpok error
	var ipAndPort string

	for ipAndPort = range chStrIPPort {

		conn = nil

		conn, tcpok = net.DialTimeout("tcp", ipAndPort, time.Duration(*connTimeOut)*time.Millisecond)
		if conn != nil && tcpok == nil {

			fmt.Println(strings.Replace(ipAndPort, ":", ",", 1))
			conn.Close()
		}

		//尝试加密连接

		//		connTls = nil
		//		connTls, tlsok = tls.Dial("tcp", ipAndPort, conf)

		//		if connTls != nil && tlsok == nil {

		//			fmt.Println(strings.Replace(ipAndPort, ":", ",", 1), ",", "tls")
		//			connTls.Close()
		//		}
		//fmt.Println("scan a port end ", ipAndPort)
	}

}
