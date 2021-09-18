package main

import "github.com/reubenmiller/go-c8y-cli-microservice/pkg/app"

func main() {
	app := app.NewApp()
	app.Run()
}
