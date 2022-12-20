package util

import (
	_ "embed"
	"strings"
)

//go:generate bash -c "git describe --tags > version.txt"
//go:embed version.txt
var VersionString string

func GetVersionString() string {
	return strings.TrimSuffix(strings.TrimSuffix(VersionString, "\r\n"), "\n")
}
