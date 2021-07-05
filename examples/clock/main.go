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

	app.Get("/", wiredServer.CreateHTMLHandler(components.NewClock, gowired.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(wiredServer.HandleWSRequest))

	_ = app.Listen(":3000")

}
