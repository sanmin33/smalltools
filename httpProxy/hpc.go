/*加密传输的proxy，采用RC4加密，
 */
package main

import (
	"crypto/rc4"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
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
		fmt.Println("tcp连接空")
		return
	}
	defer client.Close()

	address := serverIP + ":" + serverPort
	fmt.Println("服务器地址address:", address)
	tcpaddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		log.Println("tcp地址错误", address, err)
		return
	}
	server, err := net.DialTCP("tcp", nil, tcpaddr)
	if err != nil {
		log.Println("拨号服务器失败", err)
		return
	}
	//进行转发
	go pcTos.encryptCopy(server, client) //客户端收到的是明文，编码后就成了密文并传给代理的服务端
	psToc.encryptCopy(client, server)    //代理服务端发过来的是密文，编码后就成了明文，并传给浏览器
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
