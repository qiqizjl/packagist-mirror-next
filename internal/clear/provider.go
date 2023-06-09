package clear

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os"
	"packagist-mirror-next/internal/file"
	"packagist-mirror-next/internal/store"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/types"
	"sync"
	"time"
)

type Provider struct {
	svcCtx *svc.ServiceContext
	ctx    *context.Context
	logger *logrus.Entry
}

func NewClearProvider(svcCtx *svc.ServiceContext, ctx *context.Context) *Provider {
	return &Provider{
		svcCtx: svcCtx,
		ctx:    ctx,
		logger: logrus.WithField("clear", "provider"),
	}
}

func (p *Provider) Run() error {
	nowTime := time.Now().Add(-20 * time.Minute).Unix()
	if err := p.UpdateTime(); err != nil {
		p.logger.Errorf("update time error: %s", err)
		return err
	}
	if err := p.removeFile(store.PackagistProvider, nowTime); err != nil {
		p.logger.Errorf("remove file error: %s", err)
		return err
	}
	if err := p.removeFile(store.PackagistProviderPackage, nowTime); err != nil {
		p.logger.Errorf("remove file error: %s", err)
		return err
	}
	return nil
}

func (p *Provider) removeFile(key string, nowTime int64) error {
	fileList, err := p.svcCtx.FileStore.GetFileList(key, 0, nowTime)
	if err != nil {
		p.logger.Errorf("get file list error: %s", err)
		return err
	}
	p.logger.Infof("remove file list: %d", len(fileList))
	for _, fileName := range fileList {
		p.logger.Infof("remove file: %s", fileName)
		if err := p.svcCtx.File.Metadata.Delete(fileName); err != nil {
			p.logger.Errorf("remove file error: %s,fileName:%s", err, fileName)
		}
		if err := file.Delete(file.GetURL(fileName)); err != nil {
			p.logger.Errorf("remove file error: %s,fileName:%s", err, fileName)
		}
		if err := p.svcCtx.FileStore.RemoveFile(key, fileName); err != nil {
			p.logger.Errorf("remove file error: %s,fileName:%s", err, fileName)
		}
	}
	return nil
}

func (p *Provider) UpdateTime() error {
	packagesIO, err := os.ReadFile(file.GetURL("/packages.json"))
	if err != nil {
		p.logger.Errorf("read packages.json error: %s", err)
		return err
	}
	packages := &types.PackagistPackage{}
	if err := json.Unmarshal(packagesIO, packages); err != nil {
		p.logger.Errorf("unmarshal packages.json error: %s", err)
		return err
	}
	wg := sync.WaitGroup{}
	for packageProvider := range packages.ListProvider() {
		wg.Add(1)
		go func(provider types.PackagistPackageProvider) {
			defer wg.Done()
			p.logger.Infof("update provider: %s", provider.URL)
			p.svcCtx.FileStore.UpdateSuccessTime(store.PackagistProvider, provider.URL)
			p.updateProviderPackage(provider.URL)
		}(packageProvider)
	}
	wg.Wait()
	return nil
}

func (p *Provider) updateProviderPackage(url string) error {
	providerPackageIO, err := os.ReadFile(file.GetURL(url))
	if err != nil {
		p.logger.Errorf("read provider package error: %s", err)
		return err
	}
	providerPackage := &types.PackagistProviderResp{}
	if err := json.Unmarshal(providerPackageIO, providerPackage); err != nil {
		p.logger.Errorf("unmarshal provider package error: %s", err)
		return err
	}
	for packageProvider := range providerPackage.ListPackages() {
		//p.logger.Infof("update provider package: %s", packageProvider.URL)
		err := p.svcCtx.FileStore.UpdateSuccessTime(store.PackagistProviderPackage, packageProvider.URL)
		if err != nil {
			p.logger.Errorf("update provider package error: %s", err)
		}
	}
	return nil
}
