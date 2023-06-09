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
	"packagist-mirror-next/internal/types"
	"sync"
)

func getMetadataProcess(svcCtx *svc.ServiceContext, ctx *context.Context) []Process {
	processNumber := viper.GetInt("process.metadata.process_number")
	processes := make([]Process, processNumber)
	for i := 0; i < processNumber; i++ {
		processes[i] = NewMetadataProcess(svcCtx, ctx)
	}
	return processes
}

func NewMetadataProcess(svcCtx *svc.ServiceContext, ctx *context.Context) *MetadataProcess {
	return &MetadataProcess{
		Logger:    logx.WithServiceContext(svcCtx).WithField("process", "metadata"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		waitGroup: sync.WaitGroup{},
	}
}

type MetadataProcess struct {
	Logger        *logrus.Entry
	ProcessNumber int
	svcCtx        *svc.ServiceContext
	ctx           *context.Context
	nsqConsumer   *nsqx.Consumer
	waitGroup     sync.WaitGroup
}

func (p *MetadataProcess) InitProcess() error {
	p.nsqConsumer = nsqx.NewConsumer(packagistNSQ.TopicMetadata, p.getConsumerHandler())
	return nil
}

func (p *MetadataProcess) Run() error {
	p.Logger.Debugf("start nsq consumer:%s", p.svcCtx.ProcessName)
	return p.nsqConsumer.Start()
}

func (p *MetadataProcess) Stop() error {
	p.Logger.Debugf("stop nsq consumer:%s", p.svcCtx.ProcessName)
	p.nsqConsumer.Stop()
	p.waitGroup.Wait()
	return nil
}

func (p *MetadataProcess) getConsumerHandler() nsq.HandlerFunc {
	return func(msg *nsq.Message) error {
		p.Logger.Debugf("receive message: %s", msg.Body)
		var msgBody types.NsqMetadataMessage
		if err := json.Unmarshal(msg.Body, &msgBody); err != nil {
			p.Logger.Debugf("unmarshal message error: %s", err.Error())
			return err
		}
		p.svcCtx.FileStore.StartQueue("metadata", msgBody.PackageName)
		defer p.svcCtx.FileStore.RemoveQueue("metadata", msgBody.PackageName)
		p.waitGroup.Add(1)
		defer p.waitGroup.Done()
		p.Logger.Debugf("receive message: %+v", msgBody)
		// 初始化同步器
		metadataSyncer := syncer.NewMetadata(p.ctx, p.svcCtx)
		if err := metadataSyncer.Run(msgBody.PackageName, msgBody.Action); err != nil {
			p.Logger.WithField("package_name", msgBody.PackageName).Debugf("sync metadata error: %s", err.Error())
			return err
		}
		return nil
	}
}

func (p *MetadataProcess) GetProcessName() string {
	return "metadata"
}
