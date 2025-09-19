package nsqx

import (
	"fmt"

	"github.com/nsqio/go-nsq"
	"github.com/spf13/viper"
)

type Producer struct {
	NsqProducer *nsq.Producer
}

func NewProducer() *Producer {
	p := &Producer{}
	nsqConfig := nsq.NewConfig()
	nsqConfig.MaxAttempts = 3 // 最多重试三次
	p.NsqProducer, _ = nsq.NewProducer(viper.GetString("nsq.tcp_host"), nsqConfig)
	return p
}

func (p *Producer) Publish(message BaseMessage) error {
	if p.NsqProducer == nil {
		return fmt.Errorf("nsq producer is nil")
	}
	body, err := message.Encode()
	if err != nil {
		return err
	}
	return p.NsqProducer.Publish(message.GetTopic(), body)
}

func (p *Producer) Stop() {
	if p.NsqProducer != nil {
		p.NsqProducer.Stop()
	}
}
