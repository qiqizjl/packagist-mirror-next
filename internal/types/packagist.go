package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PackagistDistInfo struct {
	URL       string `json:"url"`
	Type      string `json:"type"`
	Shasum    string `json:"shasum"`
	Reference string `json:"reference"`
}

type PackagistVersionInfo struct {
	Name    string             `json:"name"`
	Version string             `json:"version"`
	Dist    *PackagistDistInfo `json:"dist"`
}

type PackagistMetadataPackage struct {
	Packages map[string][]interface{} `json:"packages"`
}

func (m *PackagistMetadataPackage) ListVersion() chan PackagistVersionInfo {
	ch := make(chan PackagistVersionInfo)
	go func() {
		for packageName, v := range m.Packages {
			for _, data := range v {
				if data == nil {
					continue
				}
				if _, ok := data.(string); ok {
					continue
				}
				// 先json
				jsonStr, _ := json.Marshal(data)
				var versionInfo PackagistVersionInfo
				_ = json.Unmarshal(jsonStr, &versionInfo)
				if versionInfo.Dist == nil {
					continue
				}
				if versionInfo.Dist.Reference == "" {
					continue
				}
				versionInfo.Name = packageName
				ch <- versionInfo
			}

		}
		close(ch)
	}()
	return ch
}

type PackagistMetadata struct {
	Packages map[string]map[string]PackagistVersionInfo `json:"packages"`
}

func (m *PackagistMetadata) ListVersion() chan PackagistVersionInfo {
	ch := make(chan PackagistVersionInfo)
	go func() {
		for packageName, v := range m.Packages {
			for _, info := range v {
				if info.Dist == nil {
					continue
				}
				if info.Dist.Reference == "" {
					// 为空不循环
					continue
				}
				info.Name = packageName
				ch <- info
			}
		}
		close(ch)
	}()
	return ch
}

type PackagistPackageProvider struct {
	Sha256       string
	URL          string
	ProviderName string
}

type PackagistPackage struct {
	ProviderIncludes map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"provider-includes"`
}

func (p *PackagistPackage) ListProvider() chan PackagistPackageProvider {
	ch := make(chan PackagistPackageProvider)
	go func() {
		for k, v := range p.ProviderIncludes {
			ch <- PackagistPackageProvider{
				Sha256:       v.Sha256,
				URL:          strings.Replace(k, "%hash%", v.Sha256, -1),
				ProviderName: k,
			}
		}
		close(ch)
	}()
	return ch
}

type PackagistChangeListItem struct {
	Action  string `json:"type"` //因为type是关键词。。
	Package string `json:"package"`
	//Time    int    `json:"time"`
}

type PackagistChangeListResp struct {
	Actions   []PackagistChangeListItem `json:"actions"`
	Timestamp int64                     `json:"timestamp"`
}

func (p *PackagistChangeListResp) ListChangeList() chan PackagistChangeListItem {
	ch := make(chan PackagistChangeListItem)
	go func() {
		for _, v := range p.Actions {
			ch <- v
		}
		close(ch)
	}()
	return ch
}

type PackagistAllPackage struct {
	PackageNames []string `json:"packageNames"`
}

type PackagistProviderResp struct {
	Providers map[string]struct {
		Sha256 string `json:"sha256"`
	} `json:"providers"`
}

func (p *PackagistProviderResp) ListPackages() chan PackagistPackageProvider {
	ch := make(chan PackagistPackageProvider)
	go func() {
		for k, v := range p.Providers {
			//"/p/%package%$%hash%.json"
			ch <- PackagistPackageProvider{
				Sha256:       v.Sha256,
				URL:          fmt.Sprintf("p/%s$%s.json", k, v.Sha256),
				ProviderName: k,
			}
		}
		close(ch)
	}()
	return ch
}
