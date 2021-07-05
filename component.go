package gowired

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html/atom"
	"golang.org/x/net/html"
)

const ComponentIdAttrKey = "go-wired-component-id"

var (
	ErrComponentNotPrepared = errors.New("Component need to be prepared")
	ErrComponentWithoutLog  = errors.New("Component without log defined")
	ErrComponentNil         = errors.New("Component nil")
)

//
type ComponentLifeTime interface {
	Create(component *WiredComponent)
	TemplateHandler(component *WiredComponent) string
	Mounted(component *WiredComponent)
	BeforeMount(component *WiredComponent)
	BeforeUnmount(component *WiredComponent)
}

type ChildWiredComponent interface{}

type ComponentContext struct {
	Pairs map[string]interface{}
}

func NewComponentContext() ComponentContext {
	return ComponentContext{
		Pairs: map[string]interface{}{},
	}
}

//
type WiredComponent struct {
	Name string

	IsMounted bool
	IsCreated bool
	Exited    bool

	log       Log
	life      *ComponentLifeCycle
	component ComponentLifeTime
	renderer  WiredRenderer

	children []*WiredComponent

	Context ComponentContext
}

// NewWiredComponent ...
func NewWiredComponent(name string, component ComponentLifeTime) *WiredComponent {
	return &WiredComponent{
		Name:      name,
		component: component,
		Context:   NewComponentContext(),
		renderer: WiredRenderer{
			state:      &WiredState{},
			template:   nil,
			formatters: make([]func(t string) string, 0),
		},
	}
}

func (w *WiredComponent) Create(life *ComponentLifeCycle) error {
	var err error

	w.life = life

	if w.log == nil {
		return ErrComponentWithoutLog
	}

	// The first notification, will notify
	// an Component without unique name
	w.notifyStage(WillCreate)

	w.Name = w.createUniqueName()

	// Get the template defined on Component
	ts := w.component.TemplateHandler(l)

	// Prepare the template content adding
	// gowired specific
	ts = w.addGoWiredComponentIDAttribute(ts)
	ts = w.signTemplateString(ts)

	// Generate go std template
	ct, err := w.generateTemplate(ts)

	if err != nil {
		return fmt.Errorf("generate template: %w", err)
	}

	w.renderer.setTemplate(ct, ts)

	//
	w.renderer.useFormatter(func(t string) string {
		d, _ := nodeFromString(t)
		_ = w.treatRender(d)
		t, _ = renderInnerHTML(d)
		return t
	})

	// Calling Component creation
	w.component.Create(l)

	// Creating children
	err = w.createChildren()

	if err != nil {
		return err
	}

	w.IsCreated = true

	w.notifyStage(Created)

	return err
}

func (w *WiredComponent) createChildren() error {
	var err error
	for _, child := range w.getChildrenComponents() {
		child.log = w.log
		child.Context = w.Context
		err = child.Create(w.life)
		if err != nil {
			panic(err)
		}

		w.children = append(w.children, child)
	}
	return err
}

func (w *WiredComponent) findComponentByID(id string) *WiredComponent {
	if w.Name == id {
		return w
	}

	for _, child := range w.children {
		if child.Name == id {
			return child
		}
	}

	for _, child := range w.children {
		found := child.findComponentByID(id)

		if found != nil {
			return found
		}
	}

	return nil
}

// Mount 2. the Component loading html
func (w *WiredComponent) Mount() error {

	if !w.IsCreated {
		return ErrComponentNotPrepared
	}

	w.notifyStage(WillMount)

	w.component.BeforeMount(l)

	err := w.MountChildren()

	if err != nil {
		return fmt.Errorf("mount children: %w", err)
	}

	w.component.Mounted(l)

	w.IsMounted = true

	w.notifyStage(Mounted)

	return nil
}

func (w *WiredComponent) MountChildren() error {
	w.notifyStage(WillMountChildren)
	for _, child := range w.getChildrenComponents() {
		err := child.Mount()

		if err != nil {
			return fmt.Errorf("child mount: %w", err)
		}
	}
	w.notifyStage(ChildrenMounted)
	return nil
}

// Render ...
func (w *WiredComponent) Render() (string, error) {
	w.log(LogTrace, "Render", logEx{"name": w.Name})

	if w.component == nil {
		return "", ErrComponentNil
	}

	text, _, err := w.renderer.Render(w.component)
	return text, err
}

func (w *WiredComponent) RenderChild(fn reflect.Value, _ ...reflect.Value) template.HTML {

	child, ok := fn.Interface().(*WiredComponent)

	if !ok {
		w.log(LogError, "child not a *gowired.WiredComponent", nil)
		return ""
	}

	render, err := child.Render()
	if err != nil {
		w.log(LogError, "render child: render", logEx{"error": err})
	}

	return template.HTML(render)
}

// WiredRender render a new version of the Component, and detect
// differences from the last render
// and sets the "new old" version  of render
func (w *WiredComponent) WiredRender() (*diff, error) {
	return w.renderer.WiredRender(w.component)
}

func (w *WiredComponent) Update() {
	w.notifyStage(Updated)
}

func (w *WiredComponent) UpdateWithSource(source *EventSource) {
	w.notifyStageWithSource(Updated, source)
}

// Kill ...
func (w *WiredComponent) Kill() error {

	w.KillChildren()

	w.log(LogTrace, "WillUnmount", logEx{"name": w.Name})

	w.component.BeforeUnmount(w)

	w.notifyStage(WillUnmount)

	w.Exited = true
	w.component = nil

	w.notifyStage(Unmounted)

	w.life = nil

	return nil
}

func (w *WiredComponent) KillChildren() {
	for _, child := range w.children {
		if err := child.Kill(); err != nil {
			w.log(LogError, "kill child", logEx{"name": child.Name})
		}
	}
}

// GetFieldFromPath ...
func (w *WiredComponent) GetFieldFromPath(path string) *reflect.Value {
	c := (*w).component
	v := reflect.ValueOf(c).Elem()

	for _, s := range strings.Split(path, ".") {

		if reflect.ValueOf(v).IsZero() {
			w.log(LogError, "field not found in Component", logEx{
				"Component": w.Name,
				"path":      path,
			})
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		// If it`s array this will work
		if i, err := strconv.Atoi(s); err == nil {
			v = v.Index(i)
		} else {
			v = v.FieldByName(s)
		}
	}
	return &v
}

func jsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// SetValueInPath ...
func (w *WiredComponent) SetValueInPath(value string, path string) error {

	v := w.GetFieldFromPath(path)
	n := reflect.New(v.Type())

	if v.Kind() == reflect.String {
		value = `"` + jsonEscape(value) + `"`
	}

	err := json.Unmarshal([]byte(value), n.Interface())
	if err != nil {
		return err
	}

	v.Set(n.Elem())
	return nil
}

// InvokeMethodInPath ...
func (w *WiredComponent) InvokeMethodInPath(path string, data map[string]string, domEvent *DOMEvent) error {
	m := reflect.ValueOf(w.component).MethodByName(path)
	if !m.IsValid() {
		return fmt.Errorf("not a valid function: %v", path)
	}

	// TODO: check for errors when calling
	switch m.Type().NumIn() {
	case 0:
		m.Call(nil)
	case 1:
		m.Call(
			[]reflect.Value{reflect.ValueOf(data)},
		)
	case 2:
		m.Call(
			[]reflect.Value{
				reflect.ValueOf(data),
				reflect.ValueOf(domEvent),
			},
		)
	}

	return nil
}

func (w *WiredComponent) createUniqueName() string {
	return w.Name + "_" + NewWiredID().GenerateSmall()
}

func (w *WiredComponent) getChildrenComponents() []*WiredComponent {
	components := make([]*WiredComponent, 0)
	v := reflect.ValueOf(w.component).Elem()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		wc, ok := v.Field(i).Interface().(*WiredComponent)
		if !ok {
			continue
		}

		components = append(components, wc)
	}
	return components
}

func (w *WiredComponent) notifyStage(ltu LifeTimeStage) {
	w.notifyStageWithSource(ltu, nil)
}

func (w *WiredComponent) notifyStageWithSource(ltu LifeTimeStage, source *EventSource) {
	if w.life == nil {
		w.log(LogWarn, "Component life updates channel is nil", nil)
		return
	}

	*w.life <- ComponentLifeTimeMessage{
		Stage:     ltu,
		Component: w,
		Source:    source,
	}
}

var rxTagName = regexp.MustCompile(`<([a-z0-9]+[ ]?)`)

func (w *WiredComponent) addGoWiredComponentIDAttribute(template string) string {
	found := rxTagName.FindString(template)
	if found != "" {
		replaceWith := found + ` ` + ComponentIdAttrKey + `="` + w.Name + `" `
		template = strings.Replace(template, found, replaceWith, 1)
	}
	return template
}

func (w *WiredComponent) generateTemplate(ts string) (*template.Template, error) {
	return template.New(w.Name).Funcs(template.FuncMap{
		"render": w.RenderChild,
	}).Parse(ts)
}

func (w *WiredComponent) treatRender(dom *html.Node) error {

	// Post treatment
	for _, node := range getAllChildrenRecursive(dom) {

		if goWiredInputAttr := getAttribute(node, "go-wired-input"); goWiredInputAttr != nil {
			addNodeAttribute(node, ":value", goWiredInputAttr.Val)
		}

		if valueAttr := getAttribute(node, ":value"); valueAttr != nil {
			removeNodeAttribute(node, ":value")

			cid, err := componentIDFromNode(node)

			if err != nil {
				return err
			}

			foundComponent := w.findComponentByID(cid)

			if foundComponent == nil {
				return fmt.Errorf("Component not found")
			}

			f := foundComponent.GetFieldFromPath(valueAttr.Val)

			if inputTypeAttr := getAttribute(node, "type"); inputTypeAttr != nil {
				switch inputTypeAttr.Val {
				case "checkbox":
					if f.Bool() {
						addNodeAttribute(node, "checked", "checked")
					} else {
						removeNodeAttribute(node, "checked")
					}
					break
				}
			} else if node.DataAtom == atom.Textarea {
				n, err := nodeFromString(fmt.Sprintf("%v", f))

				if n == nil || n.FirstChild == nil {
					continue
				}

				if err != nil {
					continue
				}

				child := n.FirstChild

				n.RemoveChild(child)

				node.AppendChild(child)
			} else {
				addNodeAttribute(node, "value", fmt.Sprintf("%v", f))
			}
		}

		if disabledAttr := getAttribute(node, ":disabled"); disabledAttr != nil {
			removeNodeAttribute(node, ":disabled")
			if disabledAttr.Val == "true" {
				addNodeAttribute(node, "disabled", "")
			} else {
				removeNodeAttribute(node, "disabled")
			}
		}
	}
	return nil
}

func (w *WiredComponent) signTemplateString(ts string) string {
	matches := rxTagName.FindAllStringSubmatchIndex(ts, -1)

	reverseSlice(matches)

	for _, match := range matches {
		startIndex := match[0]
		endIndex := match[1]

		startSlice := ts[:startIndex]
		endSlide := ts[endIndex:]
		matchedSlice := ts[startIndex:endIndex]

		uid := w.Name + "_" + NewWiredID().GenerateSmall()
		replaceWith := matchedSlice + ` go-wired-uid="` + uid + `" `
		ts = startSlice + replaceWith + endSlide
	}

	return ts
}

func componentIDFromNode(e *html.Node) (string, error) {
	for parent := e; parent != nil; parent = parent.Parent {
		if componentAttr := getAttribute(parent, ComponentIdAttrKey); componentAttr != nil {
			return componentAttr.Val, nil
		}
	}
	return "", fmt.Errorf("node not found")
}

func reverseSlice(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
