package remote

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/svc"
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
	l.Logger.Debugf("Get %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	return l.client.client.Do(req)

}

//func (l *PackagistRemote) GetAPI(url string, header http.Header) (*http.Response, error) {
//
//}
