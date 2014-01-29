##Proxy DNS query use TCP in go lang

苦于本地DNS污染，连github.com这种都经常解析不了。最近愈发频繁，所以写了这个程序。

- 采用多个dns地址轮询。
- dns 请求时，默认 read/write 都为 100ms 超时，实测已经足够，更长时间会导致网页访问变慢。
- 使用 TCP 做 DNS 解析，转发正常的 UDP 请求。
- go-cache 做缓存，默认一小时失效，pure go，无需安装其它组件。

另有使用 redis 做缓存的版本在 redis-cache 分支

依赖的两个库：

    go get github.com/miekg/dns
    go get github.com/pmylund/go-cache

跨平台编译后放到了我的 arm 开发板 pcDuino 上，现在又可以作为 DNS服务器 了 ^_^

    GOOS=linux GOARCH=arm go build src/dnsproxy.go

数台电脑，移动设备，平稳运行两天，正常解析。

##使用方法

支持的参数：

	dnss   = flag.String("dns", "192.168.2.1:53,8.8.8.8:53,8.8.4.4:53", "dns address, use `,` as sep")
	local  = flag.String("local", ":53", "local listen address")
	debug  = flag.Int("debug", 0, "debug level 0 1 2")
	cache  = flag.Bool("cache", true, "enable go-cache")
	expire = flag.Int64("expire", 3600, "default cache expire time")
	ipv6   = flag.Bool("6", false, "skip ipv6 record query AAAA")

build 生成 dnsproxy 文件后
执行：

    sudo ./dnsproxy

设置 dns 地址，使用 `,` 作分隔符

    sudo ./dnsproxy -dns=x.x.x.x:53,x.x.x.x:53

可以打印出 dns 查询日志

    sudo ./dnsproxy -debug=1

###Thanks
部分代码源自 [这里](https://gist.github.com/mrluanma/3722792)
