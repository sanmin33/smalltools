package main

import (
        "bufio"
        "fmt"
        "io"
        "log"
        "net"
        "net/http"
        "net/url"
        "strings"
)

func main() {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
        tcpaddr, err := net.ResolveTCPAddr("tcp4", ":8080")
        if err != nil {
                fmt.Println("开始tcp侦听出错")
                return
        }
        tcplisten, err := net.ListenTCP("tcp", tcpaddr)
        if err != nil {
                log.Panic(err)
        }

        for {
                client, err := tcplisten.AcceptTCP()
                if err != nil {
                        log.Panic(err)
                }

                go handleAHttp(client)
        }
}

func handleAHttp(client *net.TCPConn) {
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
                fmt.Println("转换request失败", err)
                return
        }
        var method, host, address string
        method = req.Method
        host = req.Host
        hostPortURL, err := url.Parse(host)
        fmt.Println("取request信息m:", method, "host:", host, "hostPortURL:", hostPortURL)
        if err != nil {
                log.Println(err)
                return
        }
        //取服务器域名（或IP）和端口号以便tcp拨号服务器
        hostPort := strings.Split(host, ":")
        if len(hostPort) < 2 {
                address = hostPort[0] + ":80"
        } else {
                address = host
        }

        fmt.Println("获得服务器地址address:", address)
        //获得了请求的host和port，就开始拨号吧
        tcpaddr, err := net.ResolveTCPAddr("tcp4", address)
        if err != nil {
                fmt.Println("tcp地址错误", address, err)
                return
        }
        server, err := net.DialTCP("tcp", nil, tcpaddr)
        if err != nil {
                fmt.Println(err)
                return
        }
        if method == "CONNECT" {
                fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
                fmt.Println("向客户端返回https完成连接信息")
        } else {
                server.Write(byteHeader[0:len(byteHeader)]) //如果不是https请求，则需要直接转发客户端请求给服务器

        }
        //进行转发
        go io.Copy(server, client)
        io.Copy(client, server)
}

//从字节流读到某一字符串为止,比如用于只从io口读http报头，报头和正文的分隔符为\r\n\r\n
func readSplitString(r *net.TCPConn, delim []byte) []byte {
        var rs []byte
        lenth := len(delim)
        fmt.Println("demim lenth:", lenth, "delim:", delim)
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
                /*
                        for k := 0; k < lenth; k++ {
                                fmt.Print(rs[len(rs)-1-k], " ")
                        }
                        fmt.Println("m=:", m)
                */
                if m == lenth {
                        return rs
                }
        }
        return rs
}
