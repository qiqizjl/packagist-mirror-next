package remote

import (
	"net/http"
	"time"
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
			DisableCompression: false,
			Proxy:              http.ProxyFromEnvironment,
		},
	}
}
