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
	"packagist-mirror-next/internal/nsq"
	"packagist-mirror-next/internal/remote"
	"packagist-mirror-next/internal/store"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/types"
)

type Provider struct {
	Logger    *logrus.Entry
	ctx       *context.Context
	svcCtx    *svc.ServiceContext
	packagist *remote.PackagistRemote
}

func NewProvider(ctx *context.Context, svcCtx *svc.ServiceContext) Provider {
	return Provider{
		Logger:    logx.WithServiceContext(svcCtx).WithField("syncer", "provider"),
		ctx:       ctx,
		svcCtx:    svcCtx,
		packagist: remote.NewPackagistRemote(ctx, svcCtx),
	}
}

func (l *Provider) Run(url string) error {

	resp, respBody, err := l.getProviderResp(url)
	if err != nil {
		l.Logger.Errorf("Get Provider Error: %s", err.Error())
		return err
	}
	for packageInfo := range respBody.ListPackages() {
		if err := l.dispatchProviderPackage(packageInfo); err != nil {
			l.Logger.Errorf("Dispatch Provider Package:%s Error: %s", packageInfo.URL, err.Error())
			return err
		}
	}
	if err := file.Store(file.GetURL(url), resp); err != nil {
		l.Logger.Errorf("Store Provider File Error: %s", err.Error())
		return err
	}
	if err := l.svcCtx.File.Metadata.PutFileContent(url, resp); err != nil {
		l.Logger.Errorf("Put Provider File Error: %s", err.Error())
		return err
	}
	if err := l.svcCtx.FileStore.MakeSuccess(store.PackagistProvider, url); err != nil {
		l.Logger.Errorf("Make Provider File Success Error: %s", err.Error())
		return err
	}
	return nil
}

func (l *Provider) getProviderResp(url string) ([]byte, *types.PackagistProviderResp, error) {
	resp, err := l.packagist.Get(url, nil)
	if err != nil {
		l.Logger.Errorf("Get Provider Error: %s", err.Error())
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Logger.Errorf("Read Provider Body Error: %s", err.Error())
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		l.Logger.Errorf("Get Provider Status Code Error: %d", resp.StatusCode)
		if err := l.svcCtx.FileStore.MakeError(store.PackagistProvider, url, resp.StatusCode); err != nil {
			l.Logger.Errorf("Make Provider File Error: %s", err.Error())
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("get Provider Status Code Error: %d", resp.StatusCode)
	}

	var respBody types.PackagistProviderResp
	if err := json.Unmarshal(body, &respBody); err != nil {
		l.Logger.Errorf("Unmarshal Provider Body Error: %s", err.Error())
		return nil, nil, err
	}
	return body, &respBody, nil
}

func (l *Provider) dispatchProviderPackage(info types.PackagistPackageProvider) error {
	isSuccess, err := l.svcCtx.FileStore.IsSuccess(store.PackagistProviderPackage, info.URL)
	if err != nil {
		l.Logger.Errorf("Is Provider Package Success Error: %s", err.Error())
		return err
	}
	if isSuccess {
		l.Logger.Debugf("Provider Package: %s is Success", info.URL)
		if err := l.svcCtx.FileStore.MakeSuccess(store.PackagistProviderPackage, info.URL); err != nil {
			l.Logger.Errorf("Make Provider Package Success Error: %s", err.Error())
			return err
		}
		return nil
	}
	return l.svcCtx.NSQ.Publish(&nsq.ProviderPackageMessage{
		URL:         info.URL,
		PackageName: info.ProviderName,
	})
}
