package util

import (
	tlsClient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

var Client tlsClient.HttpClient

func init() {
	Client, _ = tlsClient.NewHttpClient(tlsClient.NewNoopLogger(), []tlsClient.HttpClientOption{
		tlsClient.WithCookieJar(tlsClient.NewCookieJar()),
		tlsClient.WithTimeoutSeconds(600),
		tlsClient.WithClientProfile(profiles.Firefox_120),
	}...)
}
