一个向只提供API调用的老短信平台加入http调用功能的小转换程序        
主要涉及到go在windows下调用dll文件的问题，所以虽然程序脱离特定环境无法运行，做为日志性质备忘一下。        
    
golang学习-go调用windows下c写的dll    
go调用windows下调用c写的dll    
一、需求来源：有一供外来人员使用的公众wifi，原使用微信认证，后腾迅关闭了微信认证接口，拟切换为短信认证方式，已有移动短信发送平台只支持api调用，不提供http调用，而wifi认证平台只支持http发送。拟写一下值守转换程序接收wifi认证平台发来的http请求，并通过API方式向移动短信网关发送短信，完成wifi上网认证发送验证码。    
二、开发具备的要素：移动短信平台提供了.net api接口及demo，C++ API接口及C++DEMO。均为DLL方式提供，其中C++接口提供了.h，lib，dll三个文件及另一个供主DLL调用的DLL文件。dll是32位的。    
三、调用方法    
方法一、使用syscall,因为dll是32位的，go程序也需要编译为32位的，否则会提示找不到程序入口点，但我的dll调用一切都正常，就是执行结果不正常，至今没弄明白原因。调用其它dll则一切正常。    
import (    
	"fmt"    
	"syscall"    
	"unsafe"    
)    
    
func fff() {    
	dll32 := syscall.NewLazyDLL("ImApi.dll")    
	funcInit := dll32.NewProc("initWithDB")    
	ret, _, _ := funcInit.Call(StrPtr("xxxx"), StrPtr("xxx"), StrPtr("xxxx"), StrPtr("xxx"), StrPtr("xxxx"))    
	println(int32(ret))    
}    
func IntPtr(n int) uintptr {    
	return uintptr(n)    
}    
func StrPtr(s string) uintptr {    
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))    
}    
    
方法二、使用cgo进行调用。windows下使用cgo需要安装mingw,这个软件是在windows下模拟一个linux环境以运行linux下的一些软件，如gcc。需要说明的是，如果dll是32位的，则mingw也要安装32位的，否则链接的时候会提示被调用的函数不兼容。    
package main    
    
/*    
#cgo CFLAGS:  -I./    
#cgo LDFLAGS: -L${SRCDIR} -l ImApi    
#include <ImApi.h>    
#include <stdio.h>    
*/    
import "C"    
    
func main() {    
	C.initWithDB(C.CString("xxxx"), C.CString("xxxx"), C.CString("xxxx"), C.CString("xxxx"), C.CString("xxxx"))    
	C.sendSM(C.CString(to), C.CString(content), C.long(5))    
}    
    
以上代码仅为示例，并不能执行，我的程序是调用移动的一个老短信平台，一般没有环境，所以无法共享。    
上述代码 #cgo CFLAGS:  -I./  表示在当前目录下查找头文件，#cgo LDFLAGS: -L${SRCDIR} -l ImApi 表示在应用程序目录下查找库文件。    
cgo学习见：https://www.cntofu.com/book/73/ch2-cgo/ch2-02-basic.md    
mingw安装见 https://www.cnblogs.com/littlek1d/p/9459772.html    
安装完成写好代码后编译如果遇到如下提示    
error: expected '=', ',', ';', 'asm' or '__attribute__' before ' 问题：需要修改头文件.h    
我的程序是因为头文件里定义的一些参数有冲突，去掉头文件里的预定义部分，编译顺利通过。    
如果mingw是64位的，会造成链接失败提示和库不兼容问题，需要mingw和库文件都是32位。    
    
从上述情况可以看到使用cgo环境配置要麻烦些，但代码要简单些，go和c之间的数据类型对应也方便一些。    
但是需要头文件，无头文件只能用syscall     
我的程序最终使用Cgo解决。     
先参见：https://www.cntofu.com/book/73/ch2-cgo/ch2-02-basic.md    
