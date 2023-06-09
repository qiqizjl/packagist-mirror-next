package syncer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/file"
	"packagist-mirror-next/internal/remote"
	"packagist-mirror-next/internal/store"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/types"
	"time"
)

type Metadata struct {
	Logger    *logrus.Entry
	ctx       *context.Context
	svcCtx    *svc.ServiceContext
	packagist *remote.PackagistRemote
}

func NewMetadata(ctx *context.Context, svcCtx *svc.ServiceContext) Metadata {
	return Metadata{
		Logger:    logx.WithServiceContext(svcCtx).WithField("syncer", "metadata"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		packagist: remote.NewPackagistRemote(ctx, svcCtx),
	}
}

func (l *Metadata) Run(packageName string, action string) error {
	switch action {
	case "delete":
		return l.delete(packageName)
	case "update":
		return l.update(packageName)
	}
	return errors.New("unknown action")
}

func (l *Metadata) delete(packageName string) error {
	//_ := l.svcCtx.F
	if err := l.deletePackage(packageName); err != nil {
		l.Logger.Errorf("Delete Package %s,err: %s", packageName, err.Error())
		return err
	}
	if err := l.deletePackage(packageName + "~dev"); err != nil {
		l.Logger.Errorf("Delete Package %s,err: %s", packageName+"~dev", err.Error())
		return err
	}
	if err := l.svcCtx.FileStore.RemoveDist(packageName); err != nil {
		l.Logger.Errorf("Remove Dist %s,err: %s", packageName, err.Error())
		return err
	}
	return nil
}

func (l *Metadata) deletePackage(packageName string) error {
	if err := l.svcCtx.File.Metadata.Delete(l.getRemoteURL(packageName)); err != nil {
		return err
	}
	if err := file.Delete(file.GetMetadata(packageName)); err != nil {
		return err
	}
	return nil
}

// update: 更新Metadata信息
func (l *Metadata) update(packageName string) error {
	startTime := time.Now()
	url := l.getRemoteURL(packageName)
	// 从远处读取数据
	resp, err := l.getRemoteInfo(url, packageName)
	if err != nil {
		l.Logger.Errorf("Get Remote Info %s,err: %s", url, err.Error())
		return err
	}
	l.Logger.Debugf("Get Remote Info %s , time:%s", url, time.Now().Sub(startTime).String())
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		l.Logger.Debugf("Metadata %s not modified", url)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		if err := l.svcCtx.FileStore.MakeError(store.PackagistError, url, resp.StatusCode); err != nil {
			l.Logger.Infof("Make Error %s,err: %s", url, err.Error())
			return err
		}
		return errors.New(fmt.Sprintf("Get Remote Info %s,err: %s", url, resp.Status))
	}

	respBody, err := io.ReadAll(resp.Body)
	l.Logger.Debugf("io ReadAll %s , time:%s", url, time.Now().Sub(startTime).String())

	if err != nil {
		return err
	}
	l.Logger.Tracef("Get remote metadata %s,resp：%s", url, string(respBody))
	// 解析Resp
	var body types.PackagistMetadataPackage
	if err := json.Unmarshal(respBody, &body); err != nil {
		l.Logger.Errorf("Unmarshal Remote Metadata %s,err: %s", url, err.Error())
		return err
	}
	if err := l.dispatchDist(body); err != nil {
		l.Logger.Errorf("Dispatch dist %s,err: %s", packageName, err.Error())
		return err
	}
	l.Logger.Debugf("dispatchDist %s , time:%s", url, time.Now().Sub(startTime).String())
	//写入本地以及远程数据
	if err := l.svcCtx.File.Metadata.PutFileContent(url, respBody); err != nil {
		l.Logger.Errorf("Put remote metadata %s,err: %s", url, err.Error())
		return err
	}
	l.Logger.Debugf("put FileSystem %s , time:%s", url, time.Now().Sub(startTime).String())

	// 写入本地
	if err := file.Store(file.GetMetadata(packageName), respBody); err != nil {
		l.Logger.Errorf("Store local metadata %s,err: %s", packageName, err.Error())
		return err
	}
	// 存储最后修改时间
	if err := l.svcCtx.FileStore.SetLastModified(packageName, resp.Header.Get("Last-Modified")); err != nil {
		l.Logger.Errorf("Set Last Modified %s,err: %s", packageName, err.Error())
		return err
	}
	// makeSuccess
	if err := l.svcCtx.FileStore.MakeSuccess(store.PackagistMetadata, url); err != nil {
		l.Logger.Errorf("Make Success %s,err: %s", url, err.Error())
		return err
	}
	// 协程更新today stat
	go func() {
		if err := l.svcCtx.FileStore.UpdateTodayStat(store.PackagistMetadata, packageName); err != nil {
			l.Logger.Errorf("Update Today Stat %s,err: %s", packageName, err.Error())
		}
	}()
	return nil
}

func (l *Metadata) getRemoteInfo(url string, packageName string) (*http.Response, error) {
	lastModified, err := l.svcCtx.FileStore.GetLastModified(packageName)
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add("If-Modified-Since", lastModified)
	return l.packagist.Get(url, header)
}

func (l *Metadata) getRemoteURL(packageName string) string {
	return fmt.Sprintf("p2/%s.json", packageName)
}

func (l *Metadata) dispatchDist(body types.PackagistMetadataPackage) error {
	for info := range body.ListVersion() {
		if info.Dist.Reference == "" {
			l.Logger.Errorf("Dist %s %s shasum is empty", info.Name, info.Version)
			continue
		}
		if err := l.svcCtx.FileStore.SetDistVersion(info.Name, info.Dist.Reference, info.Dist.URL); err != nil {
			l.Logger.Errorf("Set Dist Version %s,err: %s", info.Name, err.Error())
			return err
		}
	}

	return nil
}
