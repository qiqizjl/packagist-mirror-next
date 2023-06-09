package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Qiniu struct {
	cdnURL        string
	box           *qbox.Mac
	bucket        string
	bucketManager *storage.BucketManager
}

//type FileSystem interface {
//	GetURL(key string) string
//	PutFileContent(key string, content []byte) error
//	PutFile(key string, io io.Reader) error
//	Delete(key string) error
//}

func initQiniu(accessKey, secretKey, bucket, cdnURL string) (FileSystem, error) {
	qboxMac := qbox.NewMac(accessKey, secretKey)
	return &Qiniu{
		cdnURL:        cdnURL,
		box:           qboxMac,
		bucket:        bucket,
		bucketManager: storage.NewBucketManager(qboxMac, nil),
	}, nil
}

func (q *Qiniu) Delete(key string) error {
	return q.bucketManager.Delete(q.bucket, key)
}

func (q *Qiniu) GetURL(key string) string {
	return fmt.Sprintf("%s/%s", q.cdnURL, key)
}

func (q *Qiniu) PutFileContent(key string, content []byte) error {
	ctx, deferFun := context.WithTimeout(context.Background(), 5*time.Minute)
	defer deferFun()
	ret := storage.PutRet{}
	fileLen := len(content)
	if fileLen >= 4*1024*1024 { // 大于4M走切片
		resumeUploader := storage.NewResumeUploaderV2Ex(q.getUploadCfg(), q.getClient())
		putExtra := &storage.RputV2Extra{
			PartSize: 2 * 1024 * 1024,
			TryTimes: 2,
		}
		return resumeUploader.PutWithoutSize(ctx, &ret, q.getUploadToken(key), key, bytes.NewReader(content), putExtra)
	} else {
		// 小于4m走文件上传
		putExtra := &storage.PutExtra{
			TryTimes:           1,
			HostFreezeDuration: 5 * time.Minute,
		}
		return q.getUploader().Put(ctx, &ret, q.getUploadToken(key), key, bytes.NewReader(content), int64(fileLen), putExtra)
	}

}

func (q *Qiniu) PutFile(key string, io io.Reader) error {
	resumeUploader := storage.NewResumeUploaderV2Ex(q.getUploadCfg(), q.getClient())
	ret := storage.PutRet{}
	putExtra := &storage.RputV2Extra{}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	return resumeUploader.PutWithoutSize(ctx, &ret, q.getUploadToken(key), key, io, putExtra)
}

func (q *Qiniu) getClient() *storage.Client {
	httpClient := http.Client{
		Timeout: 1 * time.Minute,
	}
	if viper.IsSet("remote.qiniu_proxy") {
		proxy, _ := url.Parse(viper.GetString("remote.qiniu_proxy"))
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	return &storage.Client{Client: &httpClient}
}

func (q *Qiniu) getUploader() *storage.FormUploader {
	return storage.NewFormUploaderEx(q.getUploadCfg(), q.getClient())
}

func (q *Qiniu) getUploadCfg() *storage.Config {
	cfg := &storage.Config{}
	cfg.UseHTTPS = false
	//cfg.UseCdnDomains = true
	return cfg
}

func (q *Qiniu) getUploadToken(key string) string {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", q.bucket, key),
	}
	return putPolicy.UploadToken(q.box)
}

func (q *Qiniu) ListObjects(prefix string) (chan ObjectInfo, error) {
	objectInfoChan := make(chan ObjectInfo, 0)

	go func() {
		defer close(objectInfoChan)
		marker := ""

		for true {
			entries, _, nextMarker, hasNext, err := q.bucketManager.ListFiles(q.bucket, prefix, "", marker, 1000)
			if err != nil {
				return
			}
			for _, object := range entries {
				objectInfoChan <- ObjectInfo{
					Key:          object.Key,
					Size:         object.Fsize,
					LastModified: time.Unix(object.PutTime/10000000, 0),
					ETag:         object.Hash,
				}
			}
			if hasNext {
				marker = nextMarker
			} else {
				break
			}
		}
	}()
	return objectInfoChan, nil
}
