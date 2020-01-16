/*加密传输的proxy，采用RC4加密，
 */
package main

import (
	"crypto/rc4"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"log"
	"net"
	"os"
	"runtime"
)

type Rc4 struct {
	C *rc4.Cipher
}

var pwd string = "协议中的头信息和正文是采用空行分开"
var ffilename = flag.String("f", "config.json", "配置文件名")

var fileName string
var serverIP string
var localPort string
var serverPort string

func init() {

	flag.Parse()
	fileName = *ffilename
}

func main() {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("打开配置文件失败")
		return
	}
	var jsondata []byte
	var buf = make([]byte, 1)
	for n, err := f.Read(buf); err == nil && n > 0; n, err = f.Read(buf) {
		jsondata = append(jsondata, buf[0])
	}
	localPort, err = jsonparser.GetString(jsondata, "localPort")
	if err != nil {
		fmt.Println("配置文件中无服务端口号", localPort, err)
		return
	}
	localPort = ":" + localPort
	serverIP, err = jsonparser.GetString(jsondata, "serverIP")
	if err != nil {
		fmt.Println("配置文件中无服务器ip", serverIP, err)
		return
	}
	serverPort, err = jsonparser.GetString(jsondata, "serverPort")
	if err != nil {
		fmt.Println("配置文件中无服务端口号", serverPort, err)
		return
	}

	pwd, err = jsonparser.GetString(jsondata, "password")
	if err != nil {
		fmt.Println("配置文件中无密码", err)
		return
	}

	if serverIP == "" || serverPort == "" {
		fmt.Println("请输入服务器IP及端口号")
		return
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tcpaddr, err := net.ResolveTCPAddr("tcp4", localPort)
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
			log.Println("当前协程数量：", runtime.NumGoroutine())
			log.Panic(err)
		}

		log.Println("当前协程数量：", runtime.NumGoroutine())
		go handleAClientConn(client)
	}
}

func handleAClientConn(client *net.TCPConn) {

	defer client.Close()
	c1, _ := rc4.NewCipher([]byte(pwd))
	c2, _ := rc4.NewCipher([]byte(pwd))
	pcTos := &Rc4{c1}
	psToc := &Rc4{c2}

	if client == nil {
		fmt.Println("tcp连接空")
		return
	}

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
