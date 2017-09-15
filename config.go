package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
)

type sitesCfg []Site

const (
	kConfigName      = "S3PROXY_CONFIG"
	kAWSKeyName      = "S3PROXY_AWS_KEY"
	kAWSSecretName   = "S3PROXY_AWS_SECRET"
	kAWSRegionName   = "S3PROXY_AWS_REGION"
	kAWSBucketName   = "S3PROXY_AWS_BUCKET"
	kUsersName       = "S3PROXY_USERS"
	kCORSKeyName     = "S3PROXY_OPTION_CORS"
	kGzipKeyName     = "S3PROXY_OPTION_GZIP"
	kWebsiteKeyName  = "S3PROXY_OPTION_WEBSITE"
	kPrefixKeyName   = "S3PROXY_OPTION_PREFIX"
	kForceSSLKeyName = "S3PROXY_OPTION_FORCE_SSL"
	kProxiedKeyName  = "S3PROXY_OPTION_PROXIED"
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
		err = site.validateWithHost()

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

	opts := Options{
		CORS:     os.Getenv(kCORSKeyName) == "true",
		Gzip:     os.Getenv(kGzipKeyName) == "true",
		Website:  os.Getenv(kWebsiteKeyName) == "true",
		Prefix:   os.Getenv(kPrefixKeyName),
		ForceSSL: os.Getenv(kForceSSLKeyName) == "true",
		Proxied:  os.Getenv(kProxiedKeyName) == "true",
	}

	s := Site{
		AWSKey:    os.Getenv(kAWSKeyName),
		AWSSecret: os.Getenv(kAWSSecretName),
		AWSRegion: os.Getenv(kAWSRegionName),
		AWSBucket: os.Getenv(kAWSBucketName),
		Users:     users,
		Options:   opts,
	}

	err = s.validate()

	if err != nil {
		return nil, err
	} else {
		return createSiteHandler(s), nil
	}
}

func createSiteHandler(s Site) http.Handler {
	var handler http.Handler

	proxy := NewS3Proxy(s.AWSKey, s.AWSSecret, s.AWSRegion, s.AWSBucket)
	handler = NewProxyHandler(proxy, s.Options.Prefix)

	if s.Options.Website {
		cfg, err := proxy.GetWebsiteConfig()
		if err != nil {
			fmt.Printf("warning: site for bucket %s configured with "+
				"website option but received error when retrieving "+
				"website config\n\t%v", s.AWSBucket, err)
		} else {
			handler = NewWebsiteHandler(handler, cfg)
		}
	}

	if s.Options.CORS {
		handler = corsHandler(handler)
	}

	if s.Options.Gzip {
		handler = handlers.CompressHandler(handler)
	}

	if len(s.Users) > 0 {
		handler = NewBasicAuthHandler(s.Users, handler)
	} else {
		fmt.Printf("warning: site for bucket %s has no configured users\n", s.AWSBucket)
	}

	if s.Options.ForceSSL {
		handler = NewSSLRedirectHandler(handler)
	}

	if s.Options.Proxied {
		handler = handlers.ProxyHeaders(handler)
	}

	return handler
}

func corsHandler(next http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"*"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"HEAD", "GET", "OPTIONS"}),
	)(next)
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

func (s Site) validateWithHost() error {
	if s.Host == "" {
		return errors.New("Host not specified")
	}

	return s.validate()
}

func (s Site) validate() error {
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
