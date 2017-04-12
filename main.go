package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Site struct {
	Host      string  `json:"host"`
	AWSKey    string  `json:"awsKey"`
	AWSSecret string  `json:"awsSecret"`
	AWSRegion string  `json:"awsRegion"`
	AWSBucket string  `json:"awsBucket"`
	Users     []User  `json:"users"`
	Options   Options `json:"options"`
}

type User struct {
	Name     string `json:"user"`
	Password string `json:"password"`
}

type Options struct {
	CORS bool `json:"cors"`
}

func main() {
	handler, err := ConfiguredProxyHandler()
	if err != nil {
		fmt.Printf("fatal: %v\n", err)
		return
	}

	port := flag.Int("port", 8080, "Port to listen on")

	flag.Parse()

	portStr := strconv.FormatInt(int64(*port), 10)

	log.Println("s3-proxy is listening on port " + portStr)
	log.Fatal(http.ListenAndServe(":"+portStr, handler))
}
