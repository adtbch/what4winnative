package main

import (
	apiPkg "github.com/lugvitc/whats4linux/api"
	fyneui "github.com/lugvitc/whats4linux/fyneui"
)

func main() {
	api := apiPkg.New()
	app := fyneui.New(api)
	app.Run()
}
