package nsqx

import (
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/spf13/viper"
	packagistNSQ "packagist-mirror-next/internal/nsq"
	"time"
)

type Consumer struct {
	Topic       string
	nsqConsumer *nsq.Consumer
}

func NewConsumer(topic string, handler nsq.Handler) *Consumer {
	consumer := &Consumer{
		Topic: topic,
	}
	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxAttempts = 3                        // 最多重试三次
	nsqConfig.DefaultRequeueDelay = 30 * time.Second // 重试间隔30秒
	nsqConfig.MaxRequeueDelay = 1 * time.Minute      // 最大重试间隔1分钟
	consumer.nsqConsumer, _ = nsq.NewConsumer(topic, packagistNSQ.CHANNEL, nsqConfig)
	// 注入handler
	consumer.nsqConsumer.AddHandler(handler)
	return consumer
}

func (l *Consumer) Start() error {
	if l.nsqConsumer == nil {
		return fmt.Errorf("nsq consumer is nil")
	}
	return l.nsqConsumer.ConnectToNSQD(viper.GetString("nsq.tcp_host"))
}

func (l *Consumer) Stop() error {
	if l.nsqConsumer != nil {
		l.nsqConsumer.Stop()
	}
	return nil
}
