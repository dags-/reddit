package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

func main() {
	port := flag.Int("port", 8088, "server port")
	flag.Parse()

	r := routing.New()
	r.Get("/api/*", handler())

	s := fasthttp.Server{
		Handler:            r.HandleRequest,
		GetOnly:            true,
		DisableKeepalive:   true,
		ReadTimeout:        time.Duration(time.Second * 2),
		WriteTimeout:       time.Duration(time.Second * 2),
		MaxConnsPerIP:      3,
		MaxRequestsPerConn: 2,
		MaxRequestBodySize: 0,
	}

	go handleStop()

	s.ListenAndServe(fmt.Sprintf(":%v", *port))
}

func handler() func(c *routing.Context) error {
	cut := len([]byte("/api/"))
	client := &http.Client{Timeout: time.Second}
	return func(c *routing.Context) error {
		uri := c.Request.RequestURI()
		url := "https://reddit.com/" + string(uri[cut:])

		rq, e := http.NewRequest("GET", url, nil)
		if e != nil {
			return e
		}
		rq.Header.Set("User-Agent", "proxy:me.dags.reddit:1.0")

		rs, e := client.Do(rq)
		if e != nil {
			return e
		}
		defer rs.Body.Close()

		c.Response.Header.Set("Access-Control-Allow-Origin", "*")
		_, e = io.Copy(c.Response.BodyWriter(), rs.Body)

		return e
	}
}

func handleStop() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "stop" {
			os.Exit(0)
		}
	}
}
