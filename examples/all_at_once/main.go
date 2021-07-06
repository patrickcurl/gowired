package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/patrickcurl/gowired"
	components "github.com/patrickcurl/gowired/examples/components"
)

type Home struct {
	gowired.WiredComponentWrapper
	Clock  *gowired.WiredComponent
	Todo   *gowired.WiredComponent
	Slider *gowired.WiredComponent
}

func NewHome() *gowired.WiredComponent {
	return gowired.NewWiredComponent("Home", &Home{
		Clock:  components.NewClock(),
		Todo:   components.NewTodo(),
		Slider: components.NewSlider(),
	})
}

func (h *Home) Mounted(_ *gowired.WiredComponent) {
	return
}

func (h *Home) TemplateHandler(_ *gowired.WiredComponent) string {
	return `
	<div>
		{{render .Clock}}
		{{render .Todo}}
		{{render .Slider}}
	</div>
	`
}

func main() {
	app := fiber.New()
	wiredServer := gowired.NewServer()

	app.Get("/", wiredServer.CreateHTMLHandler(NewHome, gowired.PageContent{
		Lang:  "us",
		Title: "Hello world",
	}))

	app.Get("/ws", websocket.New(wiredServer.HandleWSRequest))
	// app.get("/", )
	fmt.Println(app.Listen(":3000"))

}
