package router

import (
	"packagist-mirror-next/internal/core/logx"
	"packagist-mirror-next/internal/remote"
	"packagist-mirror-next/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Dist struct {
	svcCtx *svc.ServiceContext
	Logger *logrus.Entry
}

func NewDist(svcCtx *svc.ServiceContext) *Dist {
	return &Dist{
		svcCtx: svcCtx,
		Logger: logx.WithServiceContext(svcCtx).WithField("router", "dist"),
	}
}

func (d *Dist) DistGet(ctx *gin.Context) {
	owner := ctx.Param("owner")
	repo := ctx.Param("repo")
	version := ctx.Param("version")
	packageName := owner + "/" + repo
	d.Logger.Debugf("get package: %s, version: %s", packageName, version)
	d.Logger.Debugf("Params: %s", ctx.Params)
	versionInfo, err := d.svcCtx.FileStore.GetDistVersionInfo(packageName, version)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	if versionInfo == "" {
		ctx.String(404, "Not Found")
		return
	}
	httpClient := remote.GetGithubClient()
	resp, err := httpClient.Get(versionInfo)
	if err != nil {
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
