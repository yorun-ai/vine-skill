package main

import "example.com/shop/src/server/app"

func main() {
	app.New("./seed/hub.yaml", "127.0.0.1:7399").StartAndWait()
}
