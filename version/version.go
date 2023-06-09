package version

import "fmt"

const (
	VERSION = "unknown"
	SHA256  = "000000"
)

func GetUserAgent() string {
	return fmt.Sprint("packagist-mirror-next/", VERSION, " (", SHA256, ")")
}

func GetVersion() string {
	return fmt.Sprintf("%s (%s)", VERSION, SHA256)
}
