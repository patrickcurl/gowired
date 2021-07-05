package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/patrickcurl/gowired"
	"github.com/patrickcurl/gowired/examples/components"
)

func main() {

	app := fiber.New()
	wiredServer := gowired.NewServer()

	loggerbsc := gowired.NewLoggerBasic()
	loggerbsc.Level = gowired.LogDebug
	wiredServer.Log = loggerbsc.Log

	app.Get("/", wiredServer.CreateHTMLHandler(components.NewTodo, gowired.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(wiredServer.HandleWSRequest))

	_ = app.Listen(":3000")
}
