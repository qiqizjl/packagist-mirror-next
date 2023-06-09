package nsqx

import (
	"encoding/json"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
)

type APIStatResponse struct {
	Topics []struct {
		TopicName string `json:"topic_name"`
		Channels  []struct {
			ChannelName   string `json:"channel_name"`
			Depth         int    `json:"depth"`
			BackendDepth  int    `json:"backend_depth"`
			InFlightCount int    `json:"in_flight_count"`
			DeferredCount int    `json:"deferred_count"`
			MessageCount  int    `json:"message_count"`
			RequeueCount  int    `json:"requeue_count"`
			TimeoutCount  int    `json:"timeout_count"`
			ClientCount   int    `json:"client_count"`
		} `json:"channels"`
		Depth        int `json:"depth"`
		BackendDepth int `json:"backend_depth"`
		MessageCount int `json:"message_count"`
		MessageBytes int `json:"message_bytes"`
	} `json:"topics"`
}

type API struct {
}

func (a *API) Stat(topic, channel string) (*APIStatResponse, error) {
	uri := url.URL{
		Scheme: "http",
		Host:   viper.GetString("nsq.api_host"),
	}
	uri.Path = "/stats"
	query := uri.Query()
	query.Add("format", "json")
	query.Add("topic", topic)
	query.Add("channel", channel)
	uri.RawQuery = query.Encode()
	resp, err := http.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var stat APIStatResponse
	err = json.NewDecoder(resp.Body).Decode(&stat)
	if err != nil {
		return nil, err
	}
	return &stat, nil
}
