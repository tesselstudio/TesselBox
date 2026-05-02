package main

import (
	"runtime"

	"github.com/tesselstudio/TesselBox/pkg/platform"
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
