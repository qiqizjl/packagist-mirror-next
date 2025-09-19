package remote

import (
	"context"
	"fmt"
	"net/http"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/svc"
	"packagist-mirror-next/version"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type PackagistRemote struct {
	*client
	Logger *logrus.Entry
	ctx    *context.Context
	svc    *svc.ServiceContext
}

func NewPackagistRemote(ctx *context.Context, svcCtx *svc.ServiceContext) *PackagistRemote {
	return &PackagistRemote{
		client: newClient(),
		Logger: logx.WithServiceContext(svcCtx),
		ctx:    ctx,
		svc:    svcCtx,
	}
}

func (l *PackagistRemote) Get(url string, header http.Header) (*http.Response, error) {
	url = fmt.Sprintf("%s%s", viper.GetString("remote.repo"), url)
	return l.get(url, header)
}

func (l *PackagistRemote) ApiGet(url string, header http.Header) (*http.Response, error) {
	url = fmt.Sprintf("%s%s", viper.GetString("remote.api_repo"), url)
	return l.get(url, header)
}

func (l *PackagistRemote) get(url string, header http.Header) (*http.Response, error) {
	l.Logger.Debugf("Get %s,header: %v", url, header)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	header.Add("User-Agent", fmt.Sprintf("%s ; contact: i#nxx.email", version.GetVersion()))
	if err != nil {
		return nil, err
	}
	req.Header = header
	return l.client.client.Do(req)
}

//func (l *PackagistRemote) GetAPI(url string, header http.Header) (*http.Response, error) {
//
//}
