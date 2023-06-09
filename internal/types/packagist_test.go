package types

import (
	"encoding/json"
	"github.com/magiconair/properties/assert"
	"os"
	"testing"
)

func TestPackagistMetadataPackage(t *testing.T) {
	// TODO
	read, err := os.ReadFile("./first-composer.json")
	if err != nil {
		t.Error(err)
		return
	}
	var packagistMetadataPackage PackagistMetadataPackage
	if err := json.Unmarshal(read, &packagistMetadataPackage); err != nil {
		t.Error(err)
		return
	}
	for versionInfo := range packagistMetadataPackage.ListVersion() {
		t.Log(versionInfo)
		assert.Equal(t, versionInfo.Name, "alex-first-composer/first-composer")
	}
}
