package remote

import (
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

type client struct {
	client http.Client
}

func newClient() *client {
	return &client{
		client: getClient(),
	}
}

func getClient() http.Client {
	return http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DisableCompression: true, // 临时关闭 Disable  Packagist 的 GZIP 缓存有问题
			Proxy:              http.ProxyFromEnvironment,
		},
	}
}

func GetGithubClient() http.Client {
	httpClient := http.Client{
		Timeout: 1 * time.Minute,
	}
	if viper.IsSet("remote.github_proxy") {
		proxy, _ := url.Parse(viper.GetString("remote.github_proxy"))
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	return httpClient
}
