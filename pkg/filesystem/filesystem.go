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

var (
	MetadataFilesystem FileSystem
	DistFilesystem     FileSystem
)

func init() {
	var err error
	MetadataFilesystem, err = initFilesystem("metadata")
	if err != nil {
		panic(err)
	}
	DistFilesystem, err = initFilesystem("dist")
	if err != nil {
		panic(err)
	}

}

func initFilesystem(bucket string) (FileSystem, error) {
	if viper.Get("file."+bucket) == nil {
		return nil, fmt.Errorf("bucket %s not found", bucket)
	}
	switch viper.GetString("file." + bucket + ".type") {
	case "oss":
		return initAliCloudOSS(
			viper.GetString("file."+bucket+".accessId"),
			viper.GetString("file."+bucket+".accessSecret"),
			viper.GetString("file."+bucket+".endpoint"),
			viper.GetString("file."+bucket+".bucket"),
			viper.GetString("file."+bucket+".cdnDomain"),
		)
	}
	return nil, fmt.Errorf("unknown filesystem type")
}
