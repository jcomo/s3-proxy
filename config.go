package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type sitesCfg []Site

const (
	kConfigName    = "S3PROXY_CONFIG"
	kAWSKeyName    = "S3PROXY_AWS_KEY"
	kAWSSecretName = "S3PROXY_AWS_SECRET"
	kAWSRegionName = "S3PROXY_AWS_REGION"
	kAWSBucketName = "S3PROXY_AWS_BUCKET"
	kUsersName     = "S3PROXY_USERS"
)

func ConfiguredProxyHandler() (http.Handler, error) {
	_, ok := os.LookupEnv(kConfigName)

	if ok {
		return createMulti()
	} else {
		return createSingle()
	}
}

func createMulti() (http.Handler, error) {
	var cfg sitesCfg
	cfgJson := os.Getenv(kConfigName)

	err := json.Unmarshal([]byte(cfgJson), &cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg) == 0 {
		return nil, errors.New("Must specify one or more configurations")
	}

	handler := NewHostDispatchingHandler()

	for i, site := range cfg {
		err = site.validate(true)

		if err != nil {
			msg := fmt.Sprintf("%v in configuration at position %d", err, i)
			return nil, errors.New(msg)
		}

		handler.HandleHost(site.Host, createSiteHandler(site))
	}

	return handler, nil
}

func createSingle() (http.Handler, error) {
	users, err := parseUsers(os.Getenv(kUsersName))
	if err != nil {
		return nil, err
	}

	s := Site{
		AWSKey:    os.Getenv(kAWSKeyName),
		AWSSecret: os.Getenv(kAWSSecretName),
		AWSRegion: os.Getenv(kAWSRegionName),
		AWSBucket: os.Getenv(kAWSBucketName),
		Users:     users,
	}

	err = s.validate(false)

	if err != nil {
		return nil, err
	} else {
		return createSiteHandler(s), nil
	}
}

func createSiteHandler(s Site) http.HandlerFunc {
	proxy := NewS3Proxy(s.AWSKey, s.AWSSecret, s.AWSRegion, s.AWSBucket)
	proxyHandler := NewProxyHandler(proxy)

	if len(s.Users) > 0 {
		return NewBasicAuthHandler(s.Users, proxyHandler)
	} else {
		fmt.Printf("warning: site for bucket %s has no configured users\n", s.AWSBucket)
		return proxyHandler
	}
}

func parseUsers(us string) ([]User, error) {
	if us == "" {
		return []User{}, nil
	}

	pairs := strings.Split(us, ",")
	users := make([]User, len(pairs))

	for i, p := range pairs {
		parts := strings.Split(p, ":")
		if len(parts) != 2 {
			msg := fmt.Sprintf("Failed to parse user %s at position %d", p, i)
			return nil, errors.New(msg)
		}

		users[i] = User{
			Name:     parts[0],
			Password: parts[1],
		}
	}

	return users, nil
}

func (s Site) validate(withHost bool) error {
	if withHost && s.Host == "" {
		return errors.New("Host not specified")
	}

	if s.AWSKey == "" {
		return errors.New("AWS Key not specified")
	}

	if s.AWSSecret == "" {
		return errors.New("AWS Secret not specified")
	}

	if s.AWSRegion == "" {
		return errors.New("AWS Region not specified")
	}

	if s.AWSBucket == "" {
		return errors.New("AWS Bucket not specified")
	}

	return nil
}
