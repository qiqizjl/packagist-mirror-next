package types

type NsqMetadataMessage struct {
	Action      string `json:"action"`
	PackageName string `json:"package_name"`
}
