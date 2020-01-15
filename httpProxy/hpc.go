/*加密传输的proxy，采用RC4加密，
*/
package main

import (
	"bufio"
	"crypto/rc4"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
)

type Rc4 struct {
	C *rc4.Cipher
}

var pwd string = "helloworld"
var fclientport = flag.String("cp", "8080", "监听端口号")
var fip = flag.String("ip", "", "服务器IP")
var fserverport = flag.String("sp", "8080", "服务器端口")

var serverIP string
var clientPort string
var serverPort string

func init() {

	flag.Parse()
	serverIP = *fip
	clientPort = ":" + *fclientport
	serverPort = *fserverport
}

func main() {
	if serverIP == "" || serverPort == "" {
		fmt.Println("请输入服务器IP及端口号")
		return
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tcpaddr, err := net.ResolveTCPAddr("tcp4", clientPort)
	if err != nil {
		fmt.Println("侦听地址错", err)
		return
	}
	tcplisten, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		fmt.Println("开始tcp侦听出错", err)
	}

	for {
		client, err := tcplisten.AcceptTCP()
		if err != nil {
			log.Println("当前协程数量：", runtime.NumGoroutine)
			log.Panic(err)
		}

		go handleAClientConn(client)
	}
}

func handleAClientConn(client *net.TCPConn) {
	c1, _ := rc4.NewCipher([]byte(pwd))
	c2, _ := rc4.NewCipher([]byte(pwd))
	pcTos := &Rc4{c1}
	psToc := &Rc4{c2}

	if client == nil {
		return
	}
	defer client.Close()
	//取客户端字节流报头存着，以便转发给服务器
	byteHeader := readSplitString(client, []byte("\r\n\r\n"))
	fmt.Println("原始报头信息：", string(byteHeader))

	//取报头字节流后解析为结构化报头，方便获取想要的信息
	bfr := bufio.NewReader(strings.NewReader(string(byteHeader)))
	req, err := http.ReadRequest(bfr)
	if err != nil {
		log.Println("转换request失败", err)
		return
	}
	var method, host, address string
	method = req.Method
	host = req.Host
	//hostPortURL, err := url.Parse(host)
	fmt.Println("取request信息m:", method, "host:", host) //, "hostPortURL:", hostPortURL)
	if err != nil {
		log.Println(err)
		return
	}
	address = serverIP + ":" + serverPort
	fmt.Println("服务器地址address:", address)
	tcpaddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		log.Println("tcp地址错误", address, err)
		return
	}
	server, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		log.Println(err)
		return
	}
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
		fmt.Println("向客户端返回https完成连接信息")
	} else {
		//server.Write(byteHeader[0:len(byteHeader)]) //如果不是https请求，则需要直接转发客户端请求给服务器
		pcTos.C.XORKeyStream(byteHeader[0:len(byteHeader)], byteHeader[0:len(byteHeader)]) //明文通过异或后就加密了
		server.Write(byteHeader[0:len(byteHeader)])                                        //把加密后的密文传给代理的服务端

	}
	//进行转发
	go pcTos.encryptCopy(server, client) //客户端收到的是明文，编码后就成了密文并传给代理的服务端
	psToc.encryptCopy(client, server)    //代理服务端发过来的是密文，编码后就成了明文，并传给浏览器
}

//从字节流读到某一字符串为止,比如用于只从io口读http报头，报头和正文的分隔符为\r\n\r\n
func readSplitString(r *net.TCPConn, delim []byte) []byte {
	var rs []byte
	lenth := len(delim)
	//fmt.Println("demim lenth:", lenth, "delim:", delim)
	curByte := make([]byte, 1)

	//先读取分隔符长度-1个字节，以避免在下面循环中每次都要判断是否读够分隔符长度的字节。
	for k := 0; k < lenth-1; k++ {
		r.Read(curByte)
		rs = append(rs, curByte[0])
	}

	//继续读后面的字节并开始进行查找是否已经接收到报头正文分隔符
	for n, err := r.Read(curByte); err == nil && n > 0; n, err = r.Read(curByte) {
		rs = append(rs, curByte[0])

		var m int
		//从后向前逐个字节比较已读字节的最后几位是否和分隔符相同
		for m = 0; m < lenth; m++ {
			tt := len(rs)
			if rs[tt-1-m] != delim[lenth-1-m] {
				break
			}
		}
		if m == lenth {
			return rs
		}
	}
	return rs
}
func (c *Rc4) encryptCopy(dst io.Writer, src io.Reader) {
	buf := make([]byte, 4096)
	var err error
	n := 0
	for n, err = src.Read(buf); err == nil && n > 0; n, err = src.Read(buf) {
		c.C.XORKeyStream(buf[:n], buf[:n])

		dst.Write(buf[:n])
	}

}
