package store

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"packagist-mirror-next/internal/core/redisx"
	"time"
)

type FileStore struct {
	Logger *logrus.Entry
	redis  *redisx.Redis
}

func NewFileStore(store *redisx.Redis) *FileStore {
	return &FileStore{
		Logger: logrus.NewEntry(logrus.StandardLogger()),
		redis:  store,
	}
}

// IsSuccess check file upload success or not
func (f *FileStore) IsSuccess(key string, fileURL string) (bool, error) {
	redisKey := key + ":success"
	return f.redis.HExists(redisKey, fileURL)
}

func (f *FileStore) GetLastModified(packageName string) (string, error) {
	return f.redis.HGet(PackagistLastModified, packageName)
}

func (f *FileStore) SetLastModified(packageName string, lastModified string) error {
	return f.redis.HSet(PackagistLastModified, packageName, lastModified)
}

func (f *FileStore) SetDistVersion(packageName, version, url string) error {
	return f.redis.HSet(
		fmt.Sprintf(PackagistDistVersion, packageName),
		version,
		url)
}

// RemoveDist 移除Dist存储
func (f *FileStore) RemoveDist(packageName string) error {
	return f.redis.Del(
		fmt.Sprintf(PackagistDistVersion, packageName),
	)
}

func (f *FileStore) GetDistVersionInfo(packageName, version string) (string, error) {
	return f.redis.HGet(fmt.Sprintf(PackagistDistVersion, packageName), version)
}

func (f *FileStore) MakeSuccess(key string, url string) error {
	go func() {
		// 异步移除错误
		f.removeError(key, url)
	}()
	if err := f.redis.HSet(key+":success", url, time.Now().Unix()); err != nil {
		return err
	}
	return f.UpdateSuccessTime(key, url)
}

func (f *FileStore) UpdateSuccessTime(key, url string) error {
	return f.redis.ZAdd(key, url, time.Now().Unix())
}

func (f *FileStore) removeError(key string, url string) {
	for _, errorCode := range errorList {
		_ = f.redis.HDel(
			fmt.Sprintf(PackagistError, key, errorCode),
			url,
		)
	}
}

func (f *FileStore) MakeError(key string, url string, errorCode int) error {
	return f.redis.HSet(fmt.Sprintf(PackagistError, key, errorCode), url, time.Now().Unix())
}

func (f *FileStore) RemoveFile(key string, url string) error {
	go func() {
		// 异步移除错误
		f.removeError(key, url)
	}()
	if err := f.redis.HDel(key+":success", url); err != nil {
		return err
	}
	return f.redis.ZRem(key, url)
}

func (f *FileStore) UpdateTodayStat(key string, url string) error {
	redisKey := fmt.Sprintf(PackagistStat, key, time.Now().Format("20060102"))
	err := f.redis.HSet(
		redisKey,
		url,
		time.Now().Unix(),
	)
	if err != nil {
		return err
	}
	return f.redis.Expire(redisKey, time.Hour*24)
}

func (f *FileStore) GetMetadataLastSyncTime() (string, error) {
	return f.redis.Get(PackagistMetadataLastSync)
}

func (f *FileStore) SetMetadataLastSyncTime(lastSyncTime string) error {
	return f.redis.Set(PackagistMetadataLastSync, lastSyncTime, time.Duration(0))
}

func (f *FileStore) GetFileList(key string, startTime, endTime int64) ([]string, error) {
	return f.redis.ZRangeByScore(key, startTime, endTime)
}

func (f *FileStore) StartQueue(key string, url string) error {
	return f.redis.HSet(fmt.Sprintf(PackagistQueueInfo, key), url, time.Now().Unix())
}
func (f *FileStore) RemoveQueue(key string, url string) error {
	return f.redis.HDel(fmt.Sprintf(PackagistQueueInfo, key), url)
}

func (f *FileStore) GetPackagesLastSyncTime() string {
	result, _ := f.redis.Get(PackagistPackagesLastSync)
	return result
}

func (f *FileStore) SetPackagesLastSyncTime(lastSyncTime string) error {
	return f.redis.Set(PackagistPackagesLastSync, lastSyncTime, time.Duration(0))
}

func (f *FileStore) GetCount(key string) int64 {
	return f.redis.HCount(key + ":success")
}

func (f *FileStore) GetTodayUpdate(key string) int64 {
	redisKey := fmt.Sprintf(PackagistStat, key, time.Now().Format("20060102"))
	return f.redis.HCount(redisKey)
}
