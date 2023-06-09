package syncer

import (
	"context"
	"encoding/json"
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
)

type ProviderPackage struct {
	Logger    *logrus.Entry
	ctx       *context.Context
	svcCtx    *svc.ServiceContext
	packagist *remote.PackagistRemote
}

func NewProviderPackage(svc *context.Context, svcCtx *svc.ServiceContext) ProviderPackage {
	return ProviderPackage{
		Logger:    logx.WithServiceContext(svcCtx).WithField("syncer", "providerPackage"),
		ctx:       svc,
		svcCtx:    svcCtx,
		packagist: remote.NewPackagistRemote(svc, svcCtx),
	}
}

func (l *ProviderPackage) Run(url, packageName string) error {
	respBody, resp, err := l.getMetadataInfo(url, packageName)
	l.Logger.Debugf("getMetadataInfo %s,err: %s", packageName, err)
	if err != nil {
		l.Logger.Errorf("Get Remote Info %s,err: %s", url, err.Error())
		return err
	}
	if resp != nil {
		for info := range resp.ListVersion() {
			if err := l.dispatchDist(info); err != nil {
				l.Logger.Errorf("Dispatch Dist %s,err: %s", packageName, err.Error())
				return err
			}
		}
	} else {
		l.Logger.Errorf("Get Remote Info %s,err: %s", url, "resp is nil")
	}
	l.Logger.Debugf("Get Remote Info %s,success", url)
	if err := file.Store(file.GetURL(url), respBody); err != nil {
		l.Logger.Errorf("Store Provider File Error: %s", err.Error())
		return err
	}
	l.Logger.Debugf("Store Provider File %s,success", url)
	if err := l.svcCtx.File.Metadata.PutFileContent(url, respBody); err != nil {
		l.Logger.Errorf("Put Provider Package File Error: %s", err.Error())
		return err
	}
	l.Logger.Debugf("Put Provider Package File %s,success", url)
	if err := l.svcCtx.FileStore.MakeSuccess(store.PackagistProviderPackage, url); err != nil {
		l.Logger.Errorf("Make Provider Package File Success Error: %s", err.Error())
		return err
	}
	go func() {
		_ = l.svcCtx.FileStore.UpdateTodayStat(store.PackagistProviderPackage, packageName)
	}()
	return nil
}

func (l *ProviderPackage) getMetadataInfo(url, PackageName string) ([]byte, *types.PackagistMetadata, error) {
	resp, err := l.packagist.Get(url, nil)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if err := l.svcCtx.FileStore.MakeError(store.PackagistProviderPackage, PackageName, resp.StatusCode); err != nil {
			l.Logger.Errorf("Make Error %s,err: %s", PackageName, err.Error())
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("get %s,status code: %d", PackageName, resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Logger.Errorf("ReadAll %s,err: %s", PackageName, err.Error())
		return nil, nil, err
	}
	var metadataResp types.PackagistMetadata
	if err := json.Unmarshal(respBody, &metadataResp); err != nil {
		l.Logger.Errorf("Unmarshal %s,err: %s", PackageName, err.Error())
		return respBody, nil, nil
	}
	return respBody, &metadataResp, nil

}

func (l *ProviderPackage) dispatchDist(versionInfo types.PackagistVersionInfo) error {
	if versionInfo.Dist.Reference == "" {
		l.Logger.Warnf("Dist Shasum is empty,packageName:%s, version: %s", versionInfo.Name, versionInfo.Version)
		return nil
	}
	return l.svcCtx.FileStore.SetDistVersion(versionInfo.Name, versionInfo.Dist.Reference, versionInfo.Dist.URL)
}
