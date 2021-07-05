package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/patrickcurl/gowired"
	components "github.com/patrickcurl/gowired/examples/components"
)

type Home struct {
	gowired.LiveComponentWrapper
	Clock  *gowired.LiveComponent
	Todo   *gowired.LiveComponent
	Slider *gowired.LiveComponent
}

func NewHome() *gowired.LiveComponent {
	return gowired.NewLiveComponent("Home", &Home{
		Clock:  components.NewClock(),
		Todo:   components.NewTodo(),
		Slider: components.NewSlider(),
	})
}

func (h *Home) Mounted(_ *gowired.LiveComponent) {
	return
}

func (h *Home) TemplateHandler(_ *gowired.LiveComponent) string {
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
	app.get("/", )
	fmt.Println(app.Listen(":3000"))

}
