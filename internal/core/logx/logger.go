package logx

import (
	"context"
	"github.com/sirupsen/logrus"
	svc "packagist-mirror-next/internal/svc"
)

func WithContext(ctx context.Context) *logrus.Entry {
	return logrus.WithContext(ctx)
}

func WithServiceContext(svcCtx *svc.ServiceContext) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{"processName": svcCtx.ProcessName})
}

func GetLogger() *logrus.Entry {
	return logrus.NewEntry(logrus.StandardLogger())
}
