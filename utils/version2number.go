package utils

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	versionPiecewiseNumber = 4
	versionNumberLength    = 4
)

func VersionToNumber(version string) int64 {

	// 替换所有非数字和.的字符
	reg := regexp.MustCompile(`[^\d\.]`)
	version = reg.ReplaceAllLiteralString(version, "")

	versions := strings.Split(version, ".")
	if len(versions) < versionPiecewiseNumber {
		versions = append(versions, make([]string, versionPiecewiseNumber-len(versions))...)
	}

	for i := range versions {
		if len(versions[i]) > versionNumberLength {
			versions[i] = versions[i][:versionNumberLength]
		} else {
			versions[i] = leftPad2Len(versions[i], "0", versionNumberLength)
		}
	}

	version = strings.Join(versions, "")
	versionNumber, _ := strconv.ParseInt(version, 10, 0)

	return versionNumber
}
