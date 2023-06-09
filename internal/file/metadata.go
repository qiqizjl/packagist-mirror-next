package file

import "fmt"

func GetMetadata(packageName string) string {
	return fmt.Sprintf("%s/p2/%s.json", getBasePath(), packageName)
}

func GetURL(url string) string {
	return fmt.Sprintf("%s/%s", getBasePath(), url)
}

func getBasePath() string {
	return "./data/composer"
}
