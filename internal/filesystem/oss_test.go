package filesystem

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func getOSSFilesystem() FileSystem {
	ossClient, _ := initAliCloudOSS(
		"hack",
		"hack",
		"oss-cn-hangzhou.aliyuncs.com",
		"nxx-composer-test",
		"https://nxx-composer-test.oss-cn-hangzhou.aliyuncs.com",
	)
	return ossClient
}

func TestAliCloudOSS_GetURL(t *testing.T) {
	assert.Equal(t, "https://nxx-composer-test.oss-cn-hangzhou.aliyuncs.com/test.txt", getOSSFilesystem().GetURL("test.txt"))
}
