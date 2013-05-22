package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/miekg/dns"
	"log"
	"runtime"
	"time"
)

var (
	dns1   = flag.String("dns1", "202.102.134.68:53", "remote dns address")
	dns2   = flag.String("dns2", "202.102.128.68:53", "remote dns address")
	dns3   = flag.String("dns3", "8.8.8.8:53", "remote dns address")
	dns4   = flag.String("dns4", "8.8.4.4:53", "remote dns address")
	local  = flag.String("local", ":53", "local listen address")
	quiet  = flag.Bool("quiet", true, "print query and result")
	cache  = flag.Bool("cache", true, "enable redis cache")
	host   = flag.String("host", ":6379", "redis host")
	expire = flag.Int64("expire", 3600, "redis cache expire time")

	client *dns.Client
	conn   redis.Conn
)

func toMd5(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

func init() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU()*2 - 1)

	client = new(dns.Client)
	client.Net = "tcp"
	client.ReadTimeout = 100 * time.Millisecond
	client.WriteTimeout = 100 * time.Millisecond

	if *cache {
		var err error
		conn, err = redis.Dial("tcp", *host)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	dns.HandleFunc(".", proxyServe)

	failure := make(chan error, 1)

	go func(failure chan error) {
		failure <- dns.ListenAndServe(*local, "tcp", nil)
	}(failure)

	go func(failure chan error) {
		failure <- dns.ListenAndServe(*local, "udp", nil)
	}(failure)

	log.Printf("ready for accept connection on tcp/udp %s ...\n", *local)

	fmt.Println(<-failure)
}

func proxyServe(w dns.ResponseWriter, req *dns.Msg) {
	var (
		key   string
		m     *dns.Msg
		err   error
		tried int
		data  []byte
		id    uint16
		reply interface{}
	)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if req.MsgHdr.Response == true { // supposed responses sent to us are bogus
		return
	}

	id = req.Id
	req.Id = 0
	key = toMd5(req.String())
	req.Id = id

	if *cache {
		reply, _ = conn.Do("GET", key)
		if d, ok := reply.([]byte); ok && len(data) > 0 {
			data = d
			m = &dns.Msg{}
			m.Unpack(data)
			m.Id = id
			err = w.WriteMsg(m)
			goto end
		}
	}

tryResolve:
	m, _, err = client.Exchange(req, *dns1)

	if err != nil {
		m, _, err = client.Exchange(req, *dns2)

		if err != nil || len(m.Answer) == 0 {
			m, _, err = client.Exchange(req, *dns3)

			if err != nil {
				m, _, err = client.Exchange(req, *dns4)
			}
		}
	}

	if err != nil && tried < 1 {
		tried++
		goto tryResolve
	}

	if err == nil {
		data, err = m.Pack()
		if err == nil {
			_, err = w.Write(data)
			if err != nil {
				goto end
			}
		}
	}

	if err == nil {
		if *cache {
			m.Id = 0
			data, _ = m.Pack()
			conn.Do("SETEX", key, *expire, data)
			m.Id = id
		}
	}

end:
	if err != nil {
		log.Printf("failed %v %s\n", req.Question, err)
	}

	if !*quiet {
		if m != nil {
			fmt.Println(m)
		}
		fmt.Println("====================================================")
	}
}
