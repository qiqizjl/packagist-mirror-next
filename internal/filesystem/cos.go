package filesystem

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

type Cos struct {
	client *cos.Client
	cdnURL *url.URL
}

func initCos(accessId, accessSecret, endpoint, bucket, cdnDomain string) (FileSystem, error) {
	baseURL, _ := url.Parse(fmt.Sprintf("https://%s.%s", bucket, endpoint))
	cosBase := &cos.BaseURL{
		BucketURL: baseURL,
	}
	client := cos.NewClient(cosBase, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  accessId,
			SecretKey: accessSecret,
		},
	})
	cdnDomainInfo, err := url.Parse(cdnDomain)
	if err != nil {
		return nil, err
	}
	return &Cos{
		client: client,
		cdnURL: cdnDomainInfo,
	}, nil
}

func (c *Cos) PutFileContent(key string, content []byte) error {
	_, err := c.client.Object.Put(context.Background(), key, bytes.NewReader(content), nil)
	return err
}

func (c *Cos) PutFile(key string, io io.Reader) error {
	_, err := c.client.Object.Put(context.Background(), key, io, nil)
	return err
}

func (c *Cos) Delete(key string) error {
	_, err := c.client.Object.Delete(context.Background(), key)
	return err
}

func (c *Cos) GetURL(key string) string {
	result := c.cdnURL
	result.Path = path.Join(result.Path, key)
	return result.String()
}

func (c *Cos) ListObjects(prefix string) (chan ObjectInfo, error) {
	objectInfoChan := make(chan ObjectInfo, 0)

	go func() {
		defer close(objectInfoChan)
		opt := &cos.BucketGetOptions{
			Prefix:    prefix,
			Delimiter: "/",
			MaxKeys:   1000,
		}
		marker := ""

		for true {
			opt.Marker = marker
			v, _, err := c.client.Bucket.Get(context.Background(), opt)
			if err != nil {
				return
			}
			for _, content := range v.Contents {
				t, _ := time.Parse(time.RFC3339, content.LastModified)
				objectInfoChan <- ObjectInfo{
					Key:          content.Key,
					LastModified: t,
					//LastModified: content.,
					Size: content.Size,
				}
			}
			if v.IsTruncated == false {
				break
			}
			marker = v.NextMarker
		}
	}()
	return objectInfoChan, nil
}
