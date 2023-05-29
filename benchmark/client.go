
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"syscall"
	"time"
)

var (
	ip          = flag.String("ip", "127.0.0.1", "server IP")
	connections = flag.Int("conn", 10, "number of tcp connections")
	startMetric = flag.String("sm", time.Now().Format("2006-01-02T15:04:05 -0700"), "start time point of all clients")
)

func main() {
	flag.Parse()

	setLimit()

	addr := *ip + ":8899"
	log.Printf("connecting to %s", addr)
	var conns []net.Conn
	for i := 0; i < *connections; i++ {
		c, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			fmt.Println("failed to connect", i, err)
			i--
			continue
		}
		conns = append(conns, c)
	}

	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()

	log.Printf("init %d connections", len(conns))

//	tts := time.Second
//	if *connections > 100 {
//		tts = time.Millisecond * 5
//	}

	for {
		for i := 0; i < len(conns); i++ {
//			time.Sleep(tts)
			conn := conns[i]
			log.Printf("Send to %d", i)
      _, err:=conn.Write([]byte("GET / HTTP 1.1\r\n"))
      if err!=nil{
        log.Println(err)
        break
      }
		}
	}
}

func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
}
