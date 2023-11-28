## Proxy DNS query use TCP in go lang

苦于本地DNS污染，连github.com这种都经常解析不了。最近愈发频繁，所以写了这个程序。

- 采用多个dns地址轮询。
- dns 请求时，默认 read/write 都为 100ms 超时，实测已经足够，更长时间会导致网页访问变慢。
- 使用 TCP 做 DNS 解析，转发正常的 UDP 请求。
- go-cache 做缓存，默认一小时失效，pure go，无需安装其它组件。

另有使用 redis 做缓存的版本在 redis-cache 分支

Dependencies：

    go get github.com/miekg/dns
    go get github.com/pmylund/go-cache

跨平台编译后放到了我的 arm 开发板 pcDuino 上，现在又可以作为 DNS服务器 了 ^_^  
Build for special platform:

    GOOS=linux GOARCH=arm go build src/dnsproxy.go

数台电脑，移动设备，平稳运行两天，正常解析。

## Using

Supported arguments, all these could use as commandline flags like `-xxx=xxx`：

	dnss   = flag.String("dns", "192.168.2.1:53,8.8.8.8:53,8.8.4.4:53", "dns address, use `,` as sep")
	local  = flag.String("local", ":53", "local listen address")
	debug  = flag.Int("debug", 0, "debug level 0 1 2")
	cache  = flag.Bool("cache", true, "enable go-cache")
	expire = flag.Int64("expire", 3600, "default cache expire time")
	ipv6   = flag.Bool("6", false, "skip ipv6 record query AAAA")
	hostfile = flag.String("hostfile", "_output/host-file.txt, "host file for dns result intercept & substitute")

`make build` to build the executable file: `dnsproxy`

> [New Feature] To replace DNS response results using rules defined in the hostfile, you need to manually place the edited hostfile into the _output directory.

`make run` to run

Run manually：

    sudo ./dnsproxy

Set dns address，using `,` as the delimiter.

    sudo ./dnsproxy -dns=x.x.x.x:53,x.x.x.x:53

Using `-debug` flag could print the dns query log.

    sudo ./dnsproxy -debug=1

### Thanks
Part of source code from [HERE](https://gist.github.com/mrluanma/3722792)
