package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/nsq"
	"packagist-mirror-next/internal/remote"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/types"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type SyncChangeList struct {
	Logger    *logrus.Entry
	ctx       *context.Context
	svcCtx    *svc.ServiceContext
	packagist *remote.PackagistRemote
}

func NewSyncChangeList(ctx *context.Context, svcCtx *svc.ServiceContext) SyncChangeList {
	return SyncChangeList{
		Logger:    logx.WithServiceContext(svcCtx).WithField("syncer", "changeList"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		packagist: remote.NewPackagistRemote(ctx, svcCtx),
	}
}

func (l *SyncChangeList) Run() error {
	lastSyncTime, err := l.getLastSyncTime()
	if err != nil {
		l.Logger.Errorf("Get Last Sync Time Error: %s", err.Error())
		return err
	}
	// 请求API获取同步
	changeList, err := l.getChangeList(lastSyncTime)
	if err != nil {
		l.Logger.Errorf("Get Change List Error: %s", err.Error())
		return err
	}
	for item := range changeList.ListChangeList() {
		// 判断是否resync
		if item.Action == "resync" {
			if err := l.rsyncAllPackages(); err != nil {
				l.Logger.Errorf("Rsync All Packages Error: %s", err.Error())
				return err
			}
		} else {
			err = l.dispatchSyncMetadata(item.Package, item.Action)
			if err != nil {
				l.Logger.Errorf("Dispatch Sync Package %s Error: %s", item.Package, err.Error())
				return err
			}
		}
	}
	// 写入最后同步时间
	err = l.setLastSyncTime(strconv.FormatInt(changeList.Timestamp, 10))
	if err != nil {
		l.Logger.Errorf("Set Last Sync Time Error: %s", err.Error())
		return err
	}
	return nil
}

func (l *SyncChangeList) getChangeList(lastSyncTime string) (*types.PackagistChangeListResp, error) {
	//远程获取Change List
	resp, err := l.packagist.ApiGet(fmt.Sprintf("metadata/changes.json?since=%s", lastSyncTime), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		l.Logger.Errorf("Get Change List Error, StatusCode: %d", resp.StatusCode)
		return nil, errors.New(fmt.Sprintf("Get Change List Error, StatusCode: %d", resp.StatusCode))
	}
	// decode response
	var response types.PackagistChangeListResp
	err = json.NewDecoder(resp.Body).Decode(&response)
	l.Logger.Debugf("Get Change List: %+v", response)
	if err != nil {
		l.Logger.Errorf("Get Change List Success, Decode Error: %s", err.Error())
		return nil, err
	}
	return &response, nil
}

// 获得最后一次更新事件 默认兜底resync
func (l *SyncChangeList) getLastSyncTime() (string, error) {
	result, err := l.svcCtx.FileStore.GetMetadataLastSyncTime()
	if err != nil {
		return "", err
	}
	if result == "" {
		result = strconv.FormatInt(time.Now().Add(-30*24*time.Hour).UnixMilli()*10, 10)
	}
	return result, err
}

// 设置最后更新时间
func (l *SyncChangeList) setLastSyncTime(lastSyncTime string) error {
	return l.svcCtx.FileStore.SetMetadataLastSyncTime(lastSyncTime)
}

// 重新同步所有软件包
func (l *SyncChangeList) rsyncAllPackages() error {
	allPackages, err := l.getAllPackages()
	if err != nil {
		return err
	}
	for _, packageName := range allPackages {
		l.Logger.Debugf("Rsync Package %s", packageName)
		if err := l.dispatchSyncMetadata(packageName, "update"); err != nil {
			l.Logger.Errorf("dispatch Sync Package %s Error: %s", packageName, err.Error())
			return err
		}
		if err := l.dispatchSyncMetadata(packageName+"~dev", "update"); err != nil {
			l.Logger.Errorf("dispatch Sync Package %s Error: %s", packageName+"~dev", err.Error())
			return err
		}
	}
	return err
}

// 获得所有软件包
func (l *SyncChangeList) getAllPackages() ([]string, error) {
	//远程获取All Package
	resp, err := l.packagist.Get("packages/list.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Get Packages List Error, StatusCode: %d", resp.StatusCode))
	}
	var response types.PackagistAllPackage
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.PackageNames, nil
}

// 分发同步元数据
func (l *SyncChangeList) dispatchSyncMetadata(packageName string, action string) error {
	l.Logger.Debugf("Dispatch Sync Package %s Action: %s", packageName, action)
	return l.svcCtx.NSQ.Publish(&nsq.MetadataMessage{
		PackageName: packageName,
		Action:      action,
	})
}
