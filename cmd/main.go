package main

import (
	"runtime"

	"github.com/tesselstudio/TesselBox-unified/pkg/platform"
)

func init() {
	isMobile := runtime.GOOS == "android" || runtime.GOOS == "ios"
	if isMobile {
		platform.InitMobile()
	} else {
		platform.RunDesktop()
	}
}

func main() {
	select {}
}
