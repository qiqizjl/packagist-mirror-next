package process

import (
	"context"
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"packagist-mirror-next/internal/core/nsqx"
	packagistNSQ "packagist-mirror-next/internal/nsq"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/syncer"
	"sync"
)

type ProviderPackageProcess struct {
	Logger        *logrus.Entry
	ProcessNumber int
	svcCtx        *svc.ServiceContext
	ctx           *context.Context
	nsqConsumer   *nsqx.Consumer
	waitGroup     sync.WaitGroup
}

func getProviderPackageProcess(svcCtx *svc.ServiceContext, ctx *context.Context) []Process {
	processNumber := viper.GetInt("process.provider_package.process_number")
	processes := make([]Process, processNumber)
	for i := 0; i < processNumber; i++ {
		processes[i] = NewProviderPackageProcess(svcCtx, ctx)
	}
	return processes
}
func NewProviderPackageProcess(svcCtx *svc.ServiceContext, ctx *context.Context) *ProviderPackageProcess {
	return &ProviderPackageProcess{
		Logger:    logrus.WithField("process", "provider_package"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		waitGroup: sync.WaitGroup{},
	}
}

func (p *ProviderPackageProcess) InitProcess() error {
	p.nsqConsumer = nsqx.NewConsumer(packagistNSQ.TopicProviderPackage, p.getConsumerHandler())
	return nil
}

func (p *ProviderPackageProcess) Run() error {
	p.Logger.Debugf("start nsq consumer:%s", p.svcCtx.ProcessName)
	return p.nsqConsumer.Start()
}

func (p *ProviderPackageProcess) Stop() error {
	p.Logger.Debugf("stop nsq consumer:%s", p.svcCtx.ProcessName)
	p.nsqConsumer.Stop()
	p.waitGroup.Wait()
	return nil
}

func (p *ProviderPackageProcess) GetProcessName() string {
	return "provider-packages"
}

func (p *ProviderPackageProcess) getConsumerHandler() nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		p.Logger.Debugf("receive message: %s", msg.Body)
		var msgBody packagistNSQ.ProviderPackageMessage
		if err := json.Unmarshal(msg.Body, &msgBody); err != nil {
			p.Logger.Debugf("unmarshal message error: %s", err.Error())
			return err
		}
		p.svcCtx.FileStore.StartQueue("provider-package", msgBody.URL)
		defer p.svcCtx.FileStore.RemoveQueue("provider-package", msgBody.URL)
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()
		p.Logger.Debugf("receive message: %+v", msgBody)
		// 初始化同步器
		providerPackageSyncer := syncer.NewProviderPackage(p.ctx, p.svcCtx)
		if err := providerPackageSyncer.Run(msgBody.URL, msgBody.PackageName); err != nil {
			p.Logger.WithField("package", msgBody.PackageName).Debugf("sync metadata error: %s", err.Error())
			return err
		}
		return nil
	}
}
