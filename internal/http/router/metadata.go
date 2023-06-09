package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/nsq"
	"packagist-mirror-next/internal/svc"
)

type Metadata struct {
	svcCtx *svc.ServiceContext
	Logger *logrus.Entry
}

func NewMetadata(svcCtx *svc.ServiceContext) *Metadata {
	return &Metadata{
		svcCtx: svcCtx,
		Logger: logx.WithServiceContext(svcCtx).WithField("router", "metadata"),
	}
}

func (m *Metadata) RequeuePackage(ctx *gin.Context) {
	packageName := ctx.Query("packageName")
	err := m.svcCtx.NSQ.Publish(&nsq.MetadataMessage{
		PackageName: packageName,
		Action:      "update",
	})
	if err != nil {
		ctx.String(200, fmt.Sprintf("error:%s", err.Error()))
		return
	}
	ctx.String(200, "success")

}
