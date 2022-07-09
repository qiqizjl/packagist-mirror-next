/*
Copyright © 2022 SeanWang

*/
package main

import (
	"packagist-mirror-next/cmd"

	_ "packagist-mirror-next/pkg/filesystem"
	_ "packagist-mirror-next/pkg/store"
)

func main() {
	cmd.Execute()
}
