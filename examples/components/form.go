package components

import "github.com/patrickcurl/gowired"

type Form struct {
	gowired.LiveComponentWrapper
	Label        string
	DynamicInput *gowired.LiveComponent
}

func NewForm() *gowired.LiveComponent {
	return gowired.NewLiveComponent("Form", &Form{
		Label:        "",
		DynamicInput: NewDynamicInput(),
	})
}

func (d *Form) TemplateHandler(_ *gowired.LiveComponent) string {
	return `<div>
		{{ render .DynamicInput }}
	</div>`
}
