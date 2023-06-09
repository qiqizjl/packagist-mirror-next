package filesystem

import (
	"github.com/magiconair/properties/assert"
	"os"
	"testing"
)

func getQiniuFileSystem() FileSystem {
	qiniuClient, _ := initQiniu(
		os.Getenv("QINIU_ACCESS_KEY"),
		os.Getenv("QINIU_SECRET_KEY"),
		os.Getenv("QINIU_BUCKET"),
		"https://nxx-composer-metadata.nxx.com",
	)

	return qiniuClient
}

func TestQiniu_PutFileContent(t *testing.T) {
	err := getQiniuFileSystem().PutFileContent("test.json", []byte("test"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestQiniu_GetURL(t *testing.T) {
	result := getQiniuFileSystem().GetURL("%test%")
	assert.Equal(t, result, "https://nxx-composer-metadata.nxx.com/%test%")
}
