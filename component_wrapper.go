package gowired

// WiredComponentWrapper is a struct
type WiredComponentWrapper struct {
	Name      string
	Component *WiredComponent
}

func (l *WiredComponentWrapper) Create(lc *WiredComponent) {
	l.Component = lc
}

// TemplateHandler ...
func (l *WiredComponentWrapper) TemplateHandler(_ *WiredComponent) string {
	return "<div></div>"
}

// BeforeMount the Component loading html
func (l *WiredComponentWrapper) BeforeMount(_ *WiredComponent) {
}

// BeforeMount the Component loading html
func (l *WiredComponentWrapper) Mounted(_ *WiredComponent) {
}

// BeforeUnmount before we kill the Component
func (l *WiredComponentWrapper) BeforeUnmount(_ *WiredComponent) {
}

// Commit puts an boolean to the commit channel and notifies who is listening
func (l *WiredComponentWrapper) Commit() {
	l.Component.log(LogTrace, "Updated", logEx{"name": l.Component.Name})

	if l.Component.life == nil {
		l.Component.log(LogError, "call to commit on unmounted Component", logEx{"name": l.Component.Name})
		return
	}

	l.Component.Update()
}
