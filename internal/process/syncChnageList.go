package process

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/syncer"
	"sync"
	"time"
)

func getSyncChangeLogProcess(svcCtx *svc.ServiceContext, ctx *context.Context) []Process {
	processNumber := viper.GetInt("process.sync_change_list.process_number")
	processes := make([]Process, processNumber)
	for i := 0; i < processNumber; i++ {
		processes[i] = NewSyncChangeLogProcess(svcCtx, ctx)
	}
	return processes
}

func NewSyncChangeLogProcess(svcCtx *svc.ServiceContext, ctx *context.Context) *SyncChangeLogProcess {
	return &SyncChangeLogProcess{
		Logger:    logrus.WithField("process", "sync-change-list"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		waitGroup: &sync.WaitGroup{},
	}
}

type SyncChangeLogProcess struct {
	Logger        *logrus.Entry
	ProcessNumber int
	svcCtx        *svc.ServiceContext
	ctx           *context.Context
	waitGroup     *sync.WaitGroup
}

func (s *SyncChangeLogProcess) InitProcess() error {
	return nil
}
func (s *SyncChangeLogProcess) Run() error {
	go func() {
		for {
			s.waitGroup.Add(1)
			syncChange := syncer.NewSyncChangeList(s.ctx, s.svcCtx)
			if err := syncChange.Run(); err != nil {
				s.Logger.Errorf("sync change list error: %s", err.Error())
			}
			s.waitGroup.Done()
			// 1分钟遍历一次
			time.Sleep(1 * time.Minute)
		}

	}()
	return nil
}

func (s *SyncChangeLogProcess) Stop() error {
	s.waitGroup.Wait()
	return nil
}

func (s *SyncChangeLogProcess) GetProcessName() string {
	return "sync-change-list"
}
