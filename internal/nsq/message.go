package nsq

import "encoding/json"

type MetadataMessage struct {
	Action      string `json:"action"`
	PackageName string `json:"package_name"`
}

func (m *MetadataMessage) GetTopic() string {
	return TopicMetadata
}

func (m *MetadataMessage) Encode() ([]byte, error) {
	return json.Marshal(m)
}

type ProviderMessage struct {
	URL string `json:"url"`
}

func (p *ProviderMessage) GetTopic() string {
	return TopicProvider
}

func (p *ProviderMessage) Encode() ([]byte, error) {
	return json.Marshal(p)
}

type ProviderPackageMessage struct {
	URL         string `json:"url"`
	PackageName string `json:"package_name"`
}

func (p *ProviderPackageMessage) GetTopic() string {
	return TopicProviderPackage
}

func (p *ProviderPackageMessage) Encode() ([]byte, error) {
	return json.Marshal(p)
}
