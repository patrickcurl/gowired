package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	golive "github.com/patrickcurl/gowired"
	"github.com/patrickcurl/gowired/examples/components"
)

func main() {

	app := fiber.New()
	liveServer := golive.NewServer()

	loggerbsc := golive.NewLoggerBasic()
	loggerbsc.Level = golive.LogDebug
	liveServer.Log = loggerbsc.Log

	app.Get("/", liveServer.CreateHTMLHandler(components.NewTodo, golive.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(liveServer.HandleWSRequest))

	_ = app.Listen(":3000")
}
