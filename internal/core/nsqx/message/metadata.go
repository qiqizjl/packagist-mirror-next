package message

import "encoding/json"

type Metadata struct {
	Action      string `json:"action"`
	PackageName string `json:"package_name"`
}

func (m *Metadata) GetTopic() string {
	return "metadata"
}

func (m *Metadata) Encode() ([]byte, error) {
	return json.Marshal(m)
}
