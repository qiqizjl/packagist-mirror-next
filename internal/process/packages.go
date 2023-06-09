package process

import (
	"context"
	"github.com/serkanalgur/phpfuncs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"packagist-mirror-next/internal/core/nsqx"
	"packagist-mirror-next/internal/nsq"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/syncer"
	"sync"
	"time"
)

type PackagesProcess struct {
	Logger        *logrus.Entry
	ProcessNumber int
	svcCtx        *svc.ServiceContext
	ctx           *context.Context
	syncer        syncer.Packages
	waitGroup     *sync.WaitGroup
}

func getPackageProcess(svcCtx *svc.ServiceContext, ctx *context.Context) []Process {
	processNumber := viper.GetInt("process.packages.process_number")
	processes := make([]Process, processNumber)
	for i := 0; i < processNumber; i++ {
		processes[i] = NewPackagesProcess(svcCtx, ctx)
	}

	return processes
}

func NewPackagesProcess(svcCtx *svc.ServiceContext, ctx *context.Context) *PackagesProcess {
	return &PackagesProcess{
		Logger:    logrus.WithField("process", "packages"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		syncer:    syncer.NewPackages(ctx, svcCtx),
		waitGroup: &sync.WaitGroup{},
	}
}

func (p *PackagesProcess) InitProcess() error {
	return nil
}

func (p *PackagesProcess) Run() error {
	p.Logger.Debugf("start run process: %s", p.GetProcessName())
	go func() {
		if empty, _ := p.checkQueueEmpty(); !empty {
			p.Logger.Warn("queue is not empty, wait queue empty")
			if err := p.waitQueueEmpty(); err != nil {
				p.Logger.Error(err)
				return
			}
			_ = p.syncer.StorePackages()
		}
		for {
			time.Sleep(60 * time.Second)
			p.waitGroup.Add(1)
			notSync, err := p.syncer.Run()
			if err != nil {
				p.waitGroup.Done()
				p.Logger.Errorf("syncer run error: %v", err)
				continue
			}
			p.waitGroup.Done()
			if !notSync {
				p.Logger.Debugf("syncer run need sync")
				if err := p.waitQueueEmpty(); err != nil {
					p.Logger.Error(err)
					continue
				}
				if err := p.syncer.StorePackages(); err != nil {
					p.Logger.Errorf("syncer store packages error: %v", err)
				}
			}
		}

	}()
	return nil
}

func (p *PackagesProcess) Stop() error {
	p.waitGroup.Wait()
	return nil
}

func (p *PackagesProcess) GetProcessName() string {
	return "packages"
}

func (p *PackagesProcess) waitQueueEmpty() error {
	for {
		empty, err := p.checkQueueEmpty()
		if err != nil {
			p.Logger.Error(err)
			return err
		}
		if empty {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}

// checkQueueEmpty 检查队列是否为空
func (p *PackagesProcess) checkQueueEmpty() (bool, error) {
	api := nsqx.API{}
	resp, err := api.Stat("", nsq.CHANNEL)
	if err != nil {
		return false, err
	}
	for _, topic := range resp.Topics {
		if !phpfuncs.InArray(topic.TopicName, nsq.TopicPackagistWait) {
			continue
		}
		for _, topicChannel := range topic.Channels {
			if topicChannel.InFlightCount > 0 || topicChannel.Depth > 0 {
				return false, nil
			}
		}
	}
	return true, nil
}
