package components

import "github.com/patrickcurl/gowired"

type DynamicInput struct {
	gowired.LiveComponentWrapper
	Label string
}

func NewDynamicInput() *gowired.LiveComponent {
	return gowired.NewLiveComponent("DynamicInput", &DynamicInput{
		Label: "",
	})
}

func (d *DynamicInput) TemplateHandler(_ *gowired.LiveComponent) string {
	return `
		<div>
			<input type="string" go-wired-input="Label" />
			<span>{{.Label}}</span>
		</div>
	`
}
