package main

import "github.com/reubenmiller/c8y-microservice-starter/pkg/app"

func main() {
	app := app.NewApp()
	app.Run()
}
