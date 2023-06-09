package process

import (
	"context"
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/core/nsqx"
	packagistNSQ "packagist-mirror-next/internal/nsq"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/syncer"
	"sync"
)

type ProviderProcess struct {
	Logger        *logrus.Entry
	ProcessNumber int
	svcCtx        *svc.ServiceContext
	ctx           *context.Context
	nsqConsumer   *nsqx.Consumer
	waitGroup     sync.WaitGroup
}

func getProviderProcess(svcCtx *svc.ServiceContext, ctx *context.Context) []Process {
	processNumber := viper.GetInt("process.provider.process_number")
	processes := make([]Process, processNumber)
	for i := 0; i < processNumber; i++ {
		processes[i] = NewProviderProcess(svcCtx, ctx)
	}
	return processes
}

func NewProviderProcess(svcCtx *svc.ServiceContext, ctx *context.Context) *ProviderProcess {
	return &ProviderProcess{
		Logger:    logx.WithServiceContext(svcCtx).WithField("process", "provider"),
		svcCtx:    svcCtx,
		ctx:       ctx,
		waitGroup: sync.WaitGroup{},
	}
}

func (p *ProviderProcess) InitProcess() error {
	p.nsqConsumer = nsqx.NewConsumer(packagistNSQ.TopicProvider, p.getConsumerHandler())
	return nil
}

func (p *ProviderProcess) Run() error {
	p.Logger.Debugf("start nsq consumer:%s", p.svcCtx.ProcessName)
	return p.nsqConsumer.Start()
}

func (p *ProviderProcess) Stop() error {
	p.Logger.Debugf("stop nsq consumer:%s", p.svcCtx.ProcessName)
	p.nsqConsumer.Stop()
	p.waitGroup.Wait()
	return nil
}

func (p *ProviderProcess) getConsumerHandler() nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		p.Logger.Debugf("receive message: %s", msg.Body)
		var msgBody packagistNSQ.ProviderMessage
		if err := json.Unmarshal(msg.Body, &msgBody); err != nil {
			p.Logger.Debugf("unmarshal message error: %s", err.Error())
			return err
		}
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()
		p.Logger.Debugf("receive message: %+v", msgBody)
		// 初始化同步器
		metadataSyncer := syncer.NewProvider(p.ctx, p.svcCtx)
		if err := metadataSyncer.Run(msgBody.URL); err != nil {
			p.Logger.WithField("url", msgBody.URL).Debugf("sync metadata error: %s", err.Error())
			return err
		}
		return nil
	}
}

func (p *ProviderProcess) GetProcessName() string {
	return "provider"
}
