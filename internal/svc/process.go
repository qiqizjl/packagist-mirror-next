package svc

import (
	"packagist-mirror-next/internal/core/nsqx"
	"packagist-mirror-next/internal/core/redisx"
	"packagist-mirror-next/internal/filesystem"
	"packagist-mirror-next/internal/store"
)

type ServiceContext struct {
	ProcessName string
	FileStore   *store.FileStore
	File        struct {
		Metadata filesystem.FileSystem
		Dist     filesystem.FileSystem
	}
	NSQ *nsqx.Producer
}

func NewServiceContext(processName string) (*ServiceContext, error) {
	metadataFile, err := filesystem.NewFilesystem("metadata")
	if err != nil {
		return nil, err
	}
	distFile, err := filesystem.NewFilesystem("dist")
	if err != nil {
		return nil, err
	}
	return &ServiceContext{
		ProcessName: processName,
		FileStore:   store.NewFileStore(redisx.New()),
		File: struct {
			Metadata filesystem.FileSystem
			Dist     filesystem.FileSystem
		}{
			Metadata: metadataFile,
			Dist:     distFile,
		},
		NSQ: nsqx.NewProducer(),
	}, nil
}
