package gowired

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type Pet struct {
	WiredComponentWrapper
	Name  string
	Age   int
	Awake bool
}

var petComponent = NewWiredComponent("pet", &Pet{
	Name: "Catdog",
	Age:  12,
})

func TestWiredComponent_GetFieldFromPath(t *testing.T) {
	field := petComponent.GetFieldFromPath("Name")

	fmt.Printf("%v", field)

	if field.String() != "Catdog" {
		t.Error("The get field should return Catdog")
	}
}

func TestWiredComponent_SetValueInPathWithString(t *testing.T) {
	err := petComponent.SetValueInPath("Dog", "Name")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Name")

	if field.String() == "Catdog" {
		t.Error("The field has not been set")
	}

	if field.String() != "Dog" {
		t.Errorf("The field has set with different value! with ->%v", field)
	}
}

func TestWiredComponent_SetValueInPathWithNumber(t *testing.T) {
	err := petComponent.SetValueInPath("10", "Age")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Age")

	if field.Int() == 12 {
		t.Error("The field has not been set")
	}

	if field.Int() != 10 {
		t.Error("The field has not been set")
	}
}

func TestWiredComponent_SetValueInPathWithBoolean(t *testing.T) {
	err := petComponent.SetValueInPath("true", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Awake")

	if !field.Bool() {
		t.Error("The field has not been set")
	}
}

func TestWiredComponent_SetValueInPathWithBoolean2(t *testing.T) {
	err := petComponent.SetValueInPath("false", "Awake")

	if err != nil {
		t.Error(err)
		return
	}

	field := petComponent.GetFieldFromPath("Awake")

	if field.Bool() {
		t.Error("The field has not been set")
	}
}

type Clock struct {
	WiredComponentWrapper
}

func NewClock() *WiredComponent {
	return NewWiredComponent("Clock", &Clock{})
}

func (c *Clock) ActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (c *Clock) Mounted(l *WiredComponent) {
	go func() {
		for {
			if l.Exited {
				return
			}
			time.Sleep(time.Second)
			c.Commit()
		}
	}()
}

func (c *Clock) TemplateHandler(_ *WiredComponent) string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}

func TestComponent_LifeCycleSequence(t *testing.T) {

	c := NewClock()

	c.log = NewLoggerBasic().Log

	lc := make(ComponentLifeCycle)

	desired := []LifeTimeStage{
		WillCreate,
		Created,
		WillMount,
		WillMountChildren,
		ChildrenMounted,
		Mounted,
		Rendered,
		Updated,
		WillUnmount,
		Unmounted,
	}

	wg := sync.WaitGroup{}

	// Test until mounted
	wg.Add(5)

	go func() {
		count := 0
		for {
			a := <-lc

			if desired[count] != a.Stage {
				t.Error("Stage not expected, expecting", desired[count], "received", a.Stage)
			}

			count++

			if a.Stage == Mounted {
				return
			}

			wg.Done()

		}
	}()

	err := c.Create(&lc)
	if err != nil {
		t.Error(err)
	}

	err = c.Mount()
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
}

type TestComp struct {
	WiredComponentWrapper
}

func (tc *TestComp) TemplateHandler(_ *WiredComponent) string {
	return `
		<div>
			<div></div>
			<div>
				<div></div>
			</div>
			<div></div>
			<div></div>
		</div>
	`
}

func TestComponent_ComponentSignTemplate(t *testing.T) {
	var err error
	c := NewWiredComponent("Test", &TestComp{})
	c.log = NewLoggerBasic().Log
	err = c.Create(nil)

	if err != nil {
		t.Error(err)
	}

	err = c.Mount()

	if err != nil {
		t.Error(err)
	}

	fmt.Println(c.renderer.templateString)
}
