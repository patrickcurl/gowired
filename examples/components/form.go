package components

import "github.com/patrickcurl/gowired"

type Form struct {
	gowired.WiredComponentWrapper
	Label        string
	DynamicInput *gowired.WiredComponent
}

func NewForm() *gowired.WiredComponent {
	return gowired.NewWiredComponent("Form", &Form{
		Label:        "",
		DynamicInput: NewDynamicInput(),
	})
}

func (d *Form) TemplateHandler(_ *gowired.WiredComponent) string {
	return `<div>
		{{ render .DynamicInput }}
	</div>`
}
