package file

import (
	"os"
	"path/filepath"
	"time"
)

func Store(key string, content []byte) error {
	path, _ := filepath.Split(key)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(key, content, os.ModePerm)
}

func Delete(key string) error {
	if err := os.Remove(key); err != nil {
		return err
	}
	go func() {
		// 异步移除文件夹
		path, _ := filepath.Split(key)
		info, _ := os.ReadDir(path)
		if len(info) == 0 {
			_ = os.Remove(path)
		}
	}()

	return nil
}

func Copy(source, dest string) error {
	sourceIO, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return Store(dest, sourceIO)
}

func GetModTime(url string) (time.Time, error) {
	f, err := os.Stat(url)
	if err != nil {
		return time.Now(), err
	}
	return f.ModTime(), nil
}
