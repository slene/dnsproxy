##Proxy DNS query use TCP in go lang

苦于本地DNS污染，连github.com这种都经常解析不了。最近愈发频繁，所以写了这个程序。

- 使用 TCP 做 DNS 解析，转发正常的 UDP 请求。
- redis 做缓存，默认一小时失效。

依赖的两个库

    go get github.com/miekg/dns
    go get github.com/garyburd/redigo/redis

跨平台编译后放到了我的 arm 开发板 pcDuino 上，现在又可以作为 DNS服务器 了 ^_^

    GOOS=linux GOARCH=arm go build src/dnsproxy.go

数台电脑，移动设备，平稳运行两天，正常解析。

###Thanks
部分代码源自 [这里](https://gist.github.com/mrluanma/3722792)
