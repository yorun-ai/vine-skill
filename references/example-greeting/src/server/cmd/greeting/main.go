package main

import (
	serverapp "example.com/greeting/src/server/app"
	"go.yorun.ai/vine/app/standalone"
)

func main() {
	standalone.NewWithOption[*serverapp.GreetingApp](standalone.Option{
		SQLiteFile:   "./greeting-hub.sqlite",
		SeedYAMLFile: "./seed/hub.yaml",
		DashboardURL: "http://127.0.0.1:7299/",
	}).StartAndWait()
}
