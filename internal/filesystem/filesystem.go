package filesystem

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"time"
)

type FileSystem interface {
	GetURL(key string) string
	PutFileContent(key string, content []byte) error
	PutFile(key string, io io.Reader) error
	Delete(key string) error
}

type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

func NewFilesystem(bucket string) (FileSystem, error) {
	if viper.Get("file."+bucket+".bucket") == nil {
		return nil, fmt.Errorf("bucket %s not found", bucket)
	}
	switch viper.GetString("file." + bucket + ".drive") {
	case "oss":
		return initAliCloudOSS(
			viper.GetString("file."+bucket+".access_key_id"),
			viper.GetString("file."+bucket+".access_key_secret"),
			viper.GetString("file."+bucket+".endpoint"),
			viper.GetString("file."+bucket+".bucket"),
			viper.GetString("file."+bucket+".cdn_domain"),
		)
	case "cos":
		return initCos(
			viper.GetString("file."+bucket+".access_key_id"),
			viper.GetString("file."+bucket+".access_key_secret"),
			viper.GetString("file."+bucket+".endpoint"),
			viper.GetString("file."+bucket+".bucket"),
			viper.GetString("file."+bucket+".cdn_domain"),
		)
	case "qiniu":
		return initQiniu(
			viper.GetString("file."+bucket+".access_key_id"),
			viper.GetString("file."+bucket+".access_key_secret"),
			viper.GetString("file."+bucket+".bucket"),
			viper.GetString("file."+bucket+".cdn_domain"),
		)

	}
	return nil, fmt.Errorf("unknown filesystem type")
}
