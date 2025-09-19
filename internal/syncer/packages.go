package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/file"
	"packagist-mirror-next/internal/remote"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/internal/types"
	"strings"

	"github.com/sirupsen/logrus"
)

type Packages struct {
	Logger    *logrus.Entry
	ctx       *context.Context
	svcCtx    *svc.ServiceContext
	packagist *remote.PackagistRemote
}

func NewPackages(svc *context.Context, svcCtx *svc.ServiceContext) Packages {
	return Packages{
		Logger:    logx.WithServiceContext(svcCtx).WithField("syncer", "packages"),
		ctx:       svc,
		svcCtx:    svcCtx,
		packagist: remote.NewPackagistRemote(svc, svcCtx),
	}
}

func (l *Packages) Run() (bool, error) {
	resp, _, LastEditTime, err := l.getPackages()
	if err != nil {
		l.Logger.Errorf("Get Packages Error: %s", err.Error())
		return false, err
	}
	if l.svcCtx.FileStore.GetPackagesLastSyncTime() == LastEditTime {
		l.Logger.Infof("Packages Last Sync Time is same")
		return true, nil
	}
	nwePackages, err := l.makeNewPackages(resp)
	if err != nil {
		l.Logger.Errorf("Make New Packages Error: %s", err.Error())
		return false, err
	}
	if err := l.StorePackages(nwePackages); err != nil {
		l.Logger.Errorf("Store Packages Error: %s", err.Error())
		return false, err
	}
	if err := l.svcCtx.FileStore.SetPackagesLastSyncTime(LastEditTime); err != nil {
		l.Logger.Errorf("Set Packages Last Sync Time Error: %s", err.Error())
		return false, err
	}
	return false, nil
}

func (l *Packages) getPackages() ([]byte, *types.PackagistPackage, string, error) {
	resp, err := l.packagist.ApiGet("packages.json", nil)
	if err != nil {
		l.Logger.Errorf("Get Packages Error: %s", err.Error())
		return nil, nil, "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		l.Logger.Errorf("Get Packages Error: %s", resp.Status)
		return nil, nil, "", fmt.Errorf("get Packages Error: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Logger.Errorf("Get Packages Read Body Error: %s", err.Error())
		return nil, nil, "", err
	}
	var respBody types.PackagistPackage
	if err := json.Unmarshal(body, &respBody); err != nil {
		l.Logger.Errorf("Get Packages Unmarshal Error: %s", err.Error())
		return nil, nil, "", err
	}
	return body, &respBody, resp.Header.Get("Last-Modified"), nil
}

// 生成新的Package文件
func (l *Packages) makeNewPackages(body []byte) ([]byte, error) {
	// decode
	var respBody map[string]interface{}
	if err := json.Unmarshal(body, &respBody); err != nil {
		l.Logger.Errorf("Make New Packages Unmarshal Error: %s", err.Error())
		return nil, err
	}
	// repo.packagist.org替换成配置
	respBody["metadata-url"] = strings.Replace(respBody["metadata-url"].(string), "repo.packagist.org", "repo.packagist.cloud", 1)
	// Append Tips
	respBody["info"] = "Welcome Use Packagist Mirrors. See https://repo.packagist.cloud/ for more information."
	// append mirrors url
	respBody["mirrors"] = []map[string]interface{}{
		{
			"dist-url":  l.svcCtx.File.Dist.GetURL("%package%/%reference%.%type%"),
			"preferred": true,
		},
	}
	// encode
	newBody, err := json.Marshal(respBody)
	if err != nil {
		l.Logger.Errorf("Make New Packages Marshal Error: %s", err.Error())
		return nil, err
	}
	return newBody, nil
}

func (l *Packages) StorePackages(body []byte) error {
	if err := l.svcCtx.File.Metadata.PutFileContent("packages.json", body); err != nil {
		l.Logger.Errorf("Store Packages Put File Error: %s", err.Error())
		return err
	}
	// 写入本地
	if err := file.Store(file.GetURL("packages.json"), body); err != nil {
		l.Logger.Errorf("Store Packages Store File Error: %s", err.Error())
		return err
	}
	return nil
}
