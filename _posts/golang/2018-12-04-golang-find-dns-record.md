---
layout: post
title: 怎么使用go语言反查DNS记录
category: golang
tags: golang
date: 2018-12-26T13:19:54+08:00
description: DNS记录是映射文件，这些文件与DNS服务器关联，每个域与哪个IP地址关联，它们处理发送到每个域的请求.net包包含各种方法来查找DNS记录的一般细节
---

## 简介
DNS记录是映射文件，这些文件与DNS服务器关联，每个域与哪个IP地址关联，它们处理发送到每个域的请求.net包包含各种方法来查找DNS记录的一般细节.让我们运行一些例子，收集关于DNS服务器的信息和目标域的相应记录:

## A地址 和 IPv4/IPv6
net.LookupIP（）函数接受一个字符串（domain-name）并返回一个包含主机的IPv4和IPv6地址的net.IP对象片.

```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	iprecords, _ := net.LookupIP("baidu.com")
	for _, ip := range iprecords {
		fmt.Println(ip)
	}
}
```

上述程序的输出列出了以IPv4和IPv6格式返回的baidu.com的A记录.

```shell
C:\golang\dns>go run example1.go
2a03:2880:f12f:83:face:b00c:0:25de
31.13.79.35
```

## CNAME(Canonical Name) 将域名指向另外一个域名
 CNAME本质上是绑定域和子域文本别名. net.LookupCNAME（）函数接受主机名（m.baidu.com）作为字符串，并返回给定主机的单个CNAME.
 
```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	cname, _ := net.LookupCNAME("m.baidu.com")
	fmt.Println(cname)
}
```

返回 `m.baidu.com`域名的CNAME 结果如下:

```shell
C:\golang\dns>go run example2.go
star-mini.c10r.baidu.com.
```

### PTR (pointer)

根据一个IP值,查找映射的域名值,不一定没有ip地址都回生效,DNS的IP地址可以查到

```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	ptr, err := net.LookupAddr("114.114.114.114")
		if err != nil {
    		fmt.Println(err)
    	}
	for _, ptrvalue := range ptr {
		fmt.Println(ptrvalue)
	}
}
```
查找dns的返回值如下

```
C:\golang\dns>go run example3.go
public1.114dns.com.
```
## NS 将子域名解析到其他DNS服务商解析
这些记录提供从地址到名称的反向绑定. PTR记录应与前向map完全匹配. net.LookupAddr（）函数对地址执行反向查找，并返回映射到给定地址的名称列表.

```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	nameserver, _ := net.LookupNS("baidu.com")
	for _, ns := range nameserver {
		fmt.Println(ns)
	}
}
```

结果如下

```shell
C:\golang\dns>go run example4.go
&{ns3.baidu.com.}
&{ns4.baidu.com.}
&{ns7.baidu.com.}
&{dns.baidu.com.}
&{ns2.baidu.com.}
```
## MX 将域名指向邮件服务器地址
这些记录标识可以交换电子邮件的服务器. net.LookupMX（）函数将域名作为字符串，并返回按首选项排序的MX结构片段. MX结构由主机作为字符串组成，Pref是uint16.

```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	mxrecords, _ := net.LookupMX("baidu.com")
	for _, mx := range mxrecords {
		fmt.Println(mx.Host, mx.Pref)
	}
}
```
域名（baidu.com）的输出列表MX记录

```shell
C:\golang\dns>go run example5.go
mx.maillb.baidu.com. 10
mx.n.shifen.com. 15
mx1.baidu.com. 20
jpmx.baidu.com. 20
mx50.baidu.com. 20

```

## SRV 记录提供特定服务的服务器

LookupSRV函数尝试解析给定服务，协议和域名的SRV查询. 第二个参数是“tcp”或“udp”. 返回的记录按优先级排序，并按优先级在权重内随机化.

```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	cname, srvs, err := net.LookupSRV("xmpp-server", "tcp", "golang.org")
	if err != nil {
		panic(err)
	}
 
	fmt.Printf("\ncname: %s \n\n", cname)
 
	for _, srv := range srvs {
		fmt.Printf("%v:%v:%d:%d\n", srv.Target, srv.Port, srv.Priority, srv.Weight)
	}
}
```

下面的输出演示了CNAME返回，后跟SRV记录目标，端口，优先级和由冒号分隔的权重.
```
C:\golang\dns>go run example6.go
cname: _xmpp-server._tcp.golang.org.
```
## TXT 记录 文本长度限制512,通常做SPF记录(反垃圾邮件)
此文本记录存储有关SPF的信息，该信息可以识别授权服务器以代表您的组织发送电子邮件. net.LookupTXT（）函数将域名（baidu.com）作为字符串，并返回DNS TXT记录列表作为字符串片段.

```go
package main
 
import (
	"fmt"
	"net"
)
 
func main() {
	txtrecords, _ := net.LookupTXT("baidu.com")
 
	for _, txt := range txtrecords {
		fmt.Println(txt)
	}
}
```
gmail邮箱的txt值如下

```
C:\golang\dns>go run example7.go
v=spf1 include:spf1.baidu.com include:spf2.baidu.com include:spf3.baidu.com a mx ptr -all
google-site-verification=GHb98-6msqyx_qqjGl5eRatD3QTHyVB6-xQ3gJB5UwM
```