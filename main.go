package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type User struct {
	Name     string `json:"user"`
	Password string `json:"password"`
}

type Site struct {
	Host      string `json:"host"`
	AWSKey    string `json:"awsKey"`
	AWSSecret string `json:"awsSecret"`
	AWSRegion string `json:"awsRegion"`
	AWSBucket string `json:"awsBucket"`
	Users     []User `json:"users"`
}

type sitesCfg []Site

func main() {
	f, err := ioutil.ReadFile("sites.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	var cfg sitesCfg
	err = json.Unmarshal(f, &cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	handler := NewHostDispatchingHandler()

	for _, site := range cfg {
		proxy := NewS3Proxy(site.AWSKey, site.AWSSecret, site.AWSRegion, site.AWSBucket)
		proxyHandler := NewProxyHandler(proxy)
		authHandler := NewBasicAuthHandler(site.Users, proxyHandler)
		handler.HandleHost(site.Host, authHandler)
	}

	log.Fatal(http.ListenAndServe(":8080", handler))
}
