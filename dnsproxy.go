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
	dns1     = flag.String("dns1", "202.102.134.68:53", "remote dns address")
	dns2     = flag.String("dns2", "202.102.128.68:53", "remote dns address")
	dns3     = flag.String("dns3", "8.8.8.8:53", "remote dns address")
	dns4     = flag.String("dns4", "8.8.4.4:53", "remote dns address")
	local    = flag.String("local", ":53", "local listen address")
	debug    = flag.Int("debug", 0, "debug level 0 1 2")
	cache    = flag.Bool("cache", true, "enable redis cache")
	host     = flag.String("host", ":6379", "redis host")
	poolSize = flag.Int("pools", 20, "redis pool size")
	expire   = flag.Int64("expire", 3600, "redis cache expire time")
	ipv6     = flag.Bool("6", false, "skip ipv6 record query AAAA")

	client *dns.Client

	DEBUG int
	CACHE bool

	DNS []string

	redisPool *redis.Pool
)

func toMd5(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

func init() {
	flag.Parse()

	CACHE = *cache
	DEBUG = *debug

	runtime.GOMAXPROCS(runtime.NumCPU()*2 - 1)

	client = new(dns.Client)
	client.Net = "tcp"
	client.ReadTimeout = 100 * time.Millisecond
	client.WriteTimeout = 100 * time.Millisecond

	if CACHE {
		conn, err := redis.Dial("tcp", *host)
		if err != nil {
			log.Fatalln(err)
		}
		conn.Close()

		redisPool = redis.NewPool(func() (conn redis.Conn, err error) {
			conn, err = redis.Dial("tcp", *host)
			return
		}, *poolSize)
	}

	DNS = []string{*dns1, *dns2, *dns3, *dns4}
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
		key       string
		m         *dns.Msg
		err       error
		tried     bool
		data      []byte
		id        uint16
		query     []string
		questions []dns.Question
		reply     interface{}
	)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if req.MsgHdr.Response == true { // supposed responses sent to us are bogus
		return
	}

	query = make([]string, len(req.Question))

	for i, q := range req.Question {
		if q.Qtype != dns.TypeAAAA || *ipv6 {
			questions = append(questions, q)
		}
		query[i] = fmt.Sprintf("(%s %s %s)", q.Name, dns.ClassToString[q.Qclass], dns.TypeToString[q.Qtype])
	}

	if len(questions) == 0 {
		return
	}

	req.Question = questions

	conn := redisPool.Get()

	id = req.Id
	req.Id = 0
	key = toMd5(req.String())
	req.Id = id

	if CACHE {
		reply, err = conn.Do("GET", key)
		if err != nil {
			log.Printf("redis: %s", err)
			err = nil
		}
		if d, ok := reply.([]byte); ok && len(d) > 0 {
			data = d
			m = &dns.Msg{}
			m.Unpack(data)
			m.Id = id
			err = w.WriteMsg(m)

			if DEBUG > 0 {
				log.Printf("redis: HIT %v\n", query)
			}

			goto end
		} else {
			if DEBUG > 0 {
				log.Printf("redis: MISS %v\n", query)
			}
		}
	}

	for i, dns := range DNS {
		tried = i > 0
		if DEBUG > 0 {
			if tried {
				log.Printf("try: %v %s\n", query, *dns1)
			} else {
				log.Printf("resolve: %v %s\n", query, *dns1)
			}
		}
		m, _, err = client.Exchange(req, dns)
		if err == nil && len(m.Answer) > 0 {
			break
		}
	}

	if err == nil {
		if DEBUG > 0 {
			if tried {
				if len(m.Answer) == 0 {
					log.Printf("failed: %v\n", query)
				} else {
					log.Printf("bingo: %v %s\n", query, *dns2)
				}
			}
		}
		data, err = m.Pack()
		if err == nil {
			_, err = w.Write(data)
			if err != nil {
				goto end
			}
		}
	}

	if err == nil {
		if CACHE {
			m.Id = 0
			data, _ = m.Pack()
			_, err = conn.Do("SETEX", key, *expire, data)
			if err != nil {
				log.Printf("redis: %s", err)
				err = nil
			}
			m.Id = id
			if DEBUG > 0 {
				log.Printf("redis: CACHED %v\n", query)
			}
		}
	}

end:
	if DEBUG > 1 {
		fmt.Println(req)
		if m != nil {
			fmt.Println(m)
		}
	}
	if err != nil {
		log.Printf("error: %v %s\n", query, err)
	}

	if DEBUG > 1 {
		fmt.Println("====================================================")
	}

	conn.Close()
}
