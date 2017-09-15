<h1 align="center">s3-proxy</h1>

<p align="center">
A proxying server to private buckets in S3
</p>

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

### Introduction

There are many use cases where S3 is used as an object store for objects that may be intended to be accessed publicly.
Sometimes it is a requirement that restrictions be placed on who can access those objects without using the S3 API (eg. an company internal static site).
Since AWS does not provide the tools to do this, s3-proxy was born.

s3-proxy is meant to be completely configuration driven so that no source code modification or forking is necessary.
It can be deployed to your own private servers or a platform like Heroku with ease.
It supports basic auth for the use case of deploying to a publicly accessible server, although it is recommended to deploy s3-proxy within a firewall.

### Configuration

s3-proxy can be configured to run in single or multi mode.

#### Single Mode

In single mode, a single set of S3 credentials (keys, bucket, and region) are configured and all requests made to the proxy will be forwarded to that bucket.
Single mode is configured via the following (required) environment variables.

| Name                       | Description                                                                                                                                   |
| ------------------------   | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `S3PROXY_AWS_KEY`          | An AWS access key ID with read access to the bucket                                                                                           |
| `S3PROXY_AWS_SECRET`       | An AWS secret key that has read access to the bucket                                                                                          |
| `S3PROXY_AWS_REGION`       | The region where this bucket is located (eg. us-east-1)                                                                                       |
| `S3PROXY_AWS_BUCKET`       | The name of the bucket to proxy to                                                                                                            |
| `S3PROXY_USERS`            | A comma separated list of username/password pairs (eg. user1:pass,user2:pass). If specified, basic auth will be required to access any S3 key |
| `S3PROXY_OPTION_GZIP`      | `true` to gzip responses according to value of `Accept-Encoding` header                                                                       |
| `S3PROXY_OPTION_CORS`      | `true` to include basic CORS headers in response                                                                                              |
| `S3PROXY_OPTION_WEBSITE`   | `true` if this bucket should use its S3 website configuration                                                                                 |
| `S3PROXY_OPTION_PREFIX`    | Specify a prefix to be added to each path                                                                                                     |
| `S3PROXY_OPTION_FORCE_SSL` | `true` to force all requests to https                                                                                                         |
| `S3PROXY_OPTION_PROXIED`   | `true` to indicate that this app is running behind a proxy. This takes into account proxied headers                                           |



#### Multi Mode

Multi mode multiplexes the proxy over multiple buckets via virtual hosting.
Each set of configuration must be accompanied with a host, which will be used to route the request to the proper bucket.

To use multi mode instead, specify the environment variable `S3PROXY_CONFIG` where the value is a JSON array of configurations.
Each configuration has the following schema.
The options are the same with the exception of the host field that must be specified.

```json
{
  "host": "private.yourstaticdomain.com",
  "awsKey": "<YOUR AWS KEY HERE>",
  "awsSecret": "<YOUR AWS SECRET HERE>",
  "awsRegion": "us-east-1",
  "awsBucket": "my-site-bucket",
  "users": [
    {"name": "user1", "password": "pass"},
    {"name": "user2", "password": "pass"}
  ],
  "options": {
    "gzip": true,
    "cors": false,
    "website": true
  }
}
```

Run the JSON file through a JSON compacter before setting the environment variable to eliminate newlines.

**NOTE:** Multi mode requires a bit more of an advanced DNS setup where the each host that you configure must have a CNAME record to the s3-proxy.

### A note about AWS keys

It is good practice to utilize proper user management with the keys that are deployed with s3-proxy.
Any keys are that are used for proxying should be limited to have read-only access to the S3 buckets that they intend to fetch from.
Read-only access translates to the permissions: s3:GetObject, s3:GetBucketWebsite, s3:ListBucket.
