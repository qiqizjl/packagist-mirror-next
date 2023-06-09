package remote

import (
	"github.com/spf13/viper"
	"io"
	"net/http"
	"testing"
)

func setup() {
	viper.Set("remote.repo", "https://packagist.org")
}

func TestPackagistGet(t *testing.T) {
	setup()
	header := make(http.Header)
	header.Add("Test-HOOK", "Golang")
	resp, err := PackagistGet("packages.json", header)
	if err != nil {
		t.Error(err)
	}
	respStr, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
	t.Log(string(respStr))

}
