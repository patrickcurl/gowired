package components

import "github.com/patrickcurl/gowired"

type Slider struct {
	gowired.LiveComponentWrapper
	Size float32
}

func NewSlider() *gowired.LiveComponent {
	return gowired.NewLiveComponent("Slider", &Slider{
		Size: 40,
	})
}

func (t *Slider) Size2() float32 {
	return t.Size * 2
}

func (t *Slider) Size3() float32 {
	return t.Size * t.Size * 0.3
}

func (t *Slider) TemplateHandler(_ *gowired.LiveComponent) string {
	return `
		<div>
			<input go-wired-input="Size" type="range" value="{{.Size}}"/>
			<div class="" style="background-color: black; width: {{ .Size3 }}px; height: {{ .Size2 }}px">
				<div style="background-color: red; width: {{ .Size2 }}px; height: {{.Size2}}px" >
				</div>
			</div>
		</div>
	`
}
