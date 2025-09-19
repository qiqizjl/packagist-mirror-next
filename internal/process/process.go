package process

import (
	"context"
	"packagist-mirror-next/internal/svc"
)

// Process is a wrapper for os/exec.Cmd
type Process interface {
	GetProcessName() string // Get Process Name
	InitProcess() error     // Init process
	Run() error             // Run Jobs
	Stop() error            // Stop Jobs
}

var processGroup map[string][]Process

//var allProcess []string = []string{
//	"metadata",
//	"package",
//}

// InitProcess 初始化进程组
func InitProcess(svcCtx *svc.ServiceContext, ctx *context.Context) error {
	processGroup = make(map[string][]Process)
	processGroup["package"] = getPackageProcess(svcCtx, ctx)
	//2025.2.1 开始 Provider 不在更新
	// processGroup["provider"] = getProviderProcess(svcCtx, ctx)
	// processGroup["provider-package"] = getProviderPackageProcess(svcCtx, ctx)
	processGroup["sync-change-list"] = getSyncChangeLogProcess(svcCtx, ctx)
	processGroup["metadata"] = getMetadataProcess(svcCtx, ctx)
	return nil
}

func StartProcess() error {
	for _, processList := range processGroup {
		for _, process := range processList {
			if err := process.InitProcess(); err != nil {
				return err
			}
			if err := process.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func StopProcess() error {
	for _, processList := range processGroup {
		for _, process := range processList {
			if err := process.Stop(); err != nil {
				return err
			}
		}
	}
	return nil
}
