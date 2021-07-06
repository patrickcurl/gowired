package components

import "github.com/patrickcurl/gowired"

type DynamicInput struct {
	gowired.WiredComponentWrapper
	Label string
}

func NewDynamicInput() *gowired.WiredComponent {
	return gowired.NewWiredComponent("DynamicInput", &DynamicInput{
		Label: "",
	})
}

func (d *DynamicInput) TemplateHandler(_ *gowired.WiredComponent) string {
	return `
		<div>
			<input type="string" go-wired-input="Label" />
			<span>{{.Label}}</span>
		</div>
	`
}
