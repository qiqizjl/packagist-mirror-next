package filesystem

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"net/url"
	"path"
)

type AliCloudOSS struct {
	client *oss.Client
	bucket *oss.Bucket
	cdnURL *url.URL
}

func initAliCloudOSS(accessId, accessSecret, endpoint, bucket, cdnDomain string) (FileSystem, error) {
	client, err := oss.New(endpoint, accessId, accessSecret)
	if err != nil {
		return nil, err
	}
	ossBucket, err := client.Bucket(bucket)
	if err != nil {
		return nil, err
	}
	cdnDomainInfo, err := url.Parse(cdnDomain)
	if err != nil {
		return nil, err
	}
	return &AliCloudOSS{
		client: client,
		bucket: ossBucket,
		cdnURL: cdnDomainInfo,
	}, nil
}

func (aliOSS *AliCloudOSS) GetURL(key string) string {
	result := aliOSS.cdnURL
	result.Path = path.Join(result.Path, key)
	return result.String()
}

func (aliOSS *AliCloudOSS) PutFileContent(key string, content []byte) error {
	return aliOSS.bucket.PutObject(key, bytes.NewReader(content))
}

func (aliOSS *AliCloudOSS) PutFile(key string, io io.Reader) error {
	return aliOSS.bucket.PutObject(key, io)
}

func (aliOSS *AliCloudOSS) Delete(key string) error {
	return aliOSS.bucket.DeleteObject(key)
}

func (aliOSS *AliCloudOSS) ListObjects(prefix string) (chan ObjectInfo, error) {
	objectInfoChan := make(chan ObjectInfo, 0)
	go func() {
		continueToken := ""
		for {
			result, err := aliOSS.bucket.ListObjectsV2(
				oss.MaxKeys(1000),
				oss.ContinuationToken(continueToken),
				oss.Prefix(prefix),
			)
			if err != nil {
				return
			}
			for _, object := range result.Objects {
				objectInfoChan <- ObjectInfo{
					Key:          object.Key,
					Size:         object.Size,
					LastModified: object.LastModified,
					ETag:         object.ETag,
				}
			}
			if result.IsTruncated {
				continueToken = result.NextContinuationToken
			} else {
				break
			}
		}

		close(objectInfoChan)
	}()
	return objectInfoChan, nil

}
