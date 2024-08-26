package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// global mode
var (
	IsDebugMode bool = false
)

// version info
var (
	VERSION             = "3.0.0"
	BUILDTIME    string = "NONE"
	BUILDCOMMIT  string = "NONE"
	BUILDBRANCH  string = "NONE"
	BUILDREFNAME string = "NONE"
	GOVERSION    string
)

func init() {
	if len(GOVERSION) == 0 {
		GOVERSION = fmt.Sprintf("%s %s/%s", strings.Replace(runtime.Version(), "go", "", 1), runtime.GOOS, runtime.GOARCH)
	}
}

func Version() {
	const template = `peassets - PE executable program depends on aggregation tool
Version:     %s
Branch:      %s
Commit:      %s
Build Time:  %s
Go Version:  %s

`
	const tagTemplate = `peassets - PE executable program depends on aggregation tool
Version:     %s
Release:     %s
Commit:      %s
Build Time:  %s
Go Version:  %s

`
	if len(BUILDBRANCH) != 0 {
		fmt.Fprintf(os.Stdout, template, VERSION, BUILDBRANCH, BUILDCOMMIT, BUILDTIME, GOVERSION)
		return
	}
	fmt.Fprintf(os.Stdout, tagTemplate, VERSION, strings.TrimPrefix(BUILDREFNAME, "refs/tags/"), BUILDCOMMIT, BUILDTIME, GOVERSION)
}
func main() {

}
