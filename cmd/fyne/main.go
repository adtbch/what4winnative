package main

import (
	apiPkg "github.com/adtbch/whats4linux/api"
	fyneui "github.com/adtbch/whats4linux/fyneui"
)

func main() {
	api := apiPkg.New()
	app := fyneui.New(api)
	fyneui.ShowLogin(app)
	fyneui.Run(app)
}
