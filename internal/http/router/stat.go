package router

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"packagist-mirror-next/internal/file"
	"packagist-mirror-next/internal/store"
	"packagist-mirror-next/internal/svc"
	"time"
)

type Stat struct {
	svcCtx *svc.ServiceContext
	Logger *logrus.Entry
}

func NewStat(svcCtx *svc.ServiceContext) *Stat {
	return &Stat{
		svcCtx: svcCtx,
		Logger: logrus.WithField("router", "stat"),
	}
}

func (s *Stat) APIStat(ctx *gin.Context) {
	//Sat, 21 Jan 2023 08:25:21 GMT
	packagesLastTime, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", s.svcCtx.FileStore.GetPackagesLastSyncTime())
	packagesTime, _ := file.GetModTime(file.GetURL("packages.json"))
	ctx.JSON(200, gin.H{
		"status": "ok",
		"data": gin.H{
			"packages": gin.H{
				"last_update":      packagesLastTime.Unix(),
				"mirrors_time":     packagesTime.Unix(),
				"last_update_str":  packagesLastTime.Local().Format("2006-01-02 15:04:05"),
				"mirrors_time_str": packagesTime.Local().Format("2006-01-02 15:04:05"),
			},
			"providers": gin.H{
				"count": s.svcCtx.FileStore.GetCount(store.PackagistProvider),
			},
			"p1": gin.H{
				"count":        s.svcCtx.FileStore.GetCount(store.PackagistProviderPackage),
				"today_update": s.svcCtx.FileStore.GetTodayUpdate(store.PackagistProviderPackage),
			},
			"p2": gin.H{
				"count":        s.svcCtx.FileStore.GetCount(store.PackagistMetadata),
				"today_update": s.svcCtx.FileStore.GetTodayUpdate(store.PackagistMetadata),
			},
			//"queue":s.svcCtx.
		},
	})

}
