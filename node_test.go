package gowired

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"testing"
	"time"

	"golang.org/x/net/html"
)

type nodeTest struct {
	template  string
	patches     *patches
	component *WiredComponent
}

type nodeExpect struct {
	changeType ChangeType
	element    *html.Node
	content    *string
	attr       attrChange
	index      *int
}

type patchedComponent struct {
	WiredComponentWrapper
	testTemplate string
	Check        bool
}

var reSelectGowiredAttr = regexp.MustCompile(`[ ]?go-wired-uid="[a-zA-Z0-9_\-]+"`)

func (p *patchedComponent) TemplateHandler(_ *WiredComponent) string {
	return p.testTemplate
}

func newNodeTest(n nodeTest) nodeTest {
	dc := patchedComponent{}

	c := NewWiredComponent("testcomp", &dc)

	n.component = c
	c.log = NewLoggerBasic().Log

	dc.testTemplate = n.template

	_ = c.Create(nil)
	_ = c.Mount()

	_, _ = c.Render()

	dc.Check = true

	df, _ := c.WiredRender()

	n.patches = df

	return n
}

func (n *nodeTest) assert(expectations []nodeExpect, t *testing.T) {

	if len(n.patches.updates) != len(expectations) {
		t.Error("The number of instruction are len", len(n.patches.updates), "expected to be len", len(expectations))
	}

	for indexExpected, expected := range expectations {

		foundGiven := false

		for indexGiven, given := range n.patches.updates {

			if indexExpected == indexGiven {
				foundGiven = true
			} else {
				continue
			}

			if given.changeType != expected.changeType {
				t.Error("type is different given:", given.changeType, "expeted:", expected.changeType)
			}
			a := reSelectGowiredAttr.ReplaceAllString(given.content, "")

			if expected.content != nil && a != *expected.content {
				t.Error("contents are different given:", a, "expeted:", *expected.content)
			}

			if expected.attr != (attrChange{}) && !reflect.DeepEqual(given.attr, expected.attr) {
				t.Error("attributes are different given:", given.attr, "expeted:", expected.attr)
			}

			if !reflect.DeepEqual(pathToComponentRoot(given.element), pathToComponentRoot(expected.element)) {
				t.Error("elements with different elements given:", pathToComponentRoot(given.element), "expeted:", pathToComponentRoot(expected.element))
			}

		}

		if !foundGiven {
			t.Error("given instruction not found")
		}

	}

	if t.Failed() {
		t.Log("Time", time.Now().Format(time.Kitchen), string(debug.Stack()))
	}
}

func TestDiff_RemovedNestedText(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<h1><span>{{ if .Check }}{{else}}hello world{{ end }}</span></h1>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: SetInnerHTML,
			element:    dt.patches.actual.FirstChild.LastChild,
			attr:       attrChange{},
		},
	}, t)
}

func TestPatch_ChangeNestedText(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>Hello world<span>{{ if .Check }}hello{{ else }}hello world{{ end }}</span></div>`,
	})
	c := "hello"
	dt.assert([]nodeExpect{
		{
			changeType: SetInnerHTML,
			element:    dt.patches.actual.FirstChild.LastChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_RemoveElement(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>{{ if .Check }}{{else}}<div></div>{{ end }}</div>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: Remove,
			element:    dt.patches.actual.FirstChild.FirstChild,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AppendElement(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>{{ if .Check }}<div></div>{{else}}{{ end }}</div>`,
	})

	c := "<div></div>"
	dt.assert([]nodeExpect{
		{
			changeType: Append,
			element:    dt.patches.actual.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AppendNestedElements(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>{{ if .Check }}<div><div></div></div>{{ end }}</div>`,
	})

	c := "<div><div></div></div>"
	dt.assert([]nodeExpect{
		{
			changeType: Append,
			element:    dt.patches.actual.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ReplaceNestedElementsWithText(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>{{ if .Check }}<div>a<div>a</div></div>{{ else }}<span></span>{{end}}</div>`,
	})

	c := "<div>a<div>a</div></div>"
	dt.assert([]nodeExpect{
		{
			changeType: Replace,
			element:    dt.patches.actual.FirstChild.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_ReplaceTagWithContent(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>{{ if .Check }}<div>a</div>{{ else }}<span>a</span>{{ end }}</div>`,
	})

	c := "<div>a</div>"
	dt.assert([]nodeExpect{
		{
			changeType: Replace,
			element:    dt.patches.actual.FirstChild.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_AddAttribute(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div {{ if .Check }}disabled{{ end }}></div>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "",
			},
		},
	}, t)
}

func TestDiff_RemoveAttribute(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div {{ if not .Check }}disabled{{ end }}></div>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: RemoveAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "",
			},
		},
	}, t)
}

func TestDiff_AddTextContent(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<div>{{ if .Check }}aaaa{{ end }}</div>`,
	})

	c := "aaaa"
	dt.assert([]nodeExpect{
		{
			changeType: SetInnerHTML,
			element:    dt.patches.actual.FirstChild,
			content:    &c,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffWithTabs(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `{{ if .Check }}
	<div></div>
{{else}}
<div></div>
{{end}}`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: Replace,
			element:    dt.patches.actual.FirstChild.NextSibling,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffWithTabsAndBreakLine(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `{{ if .Check }}
	<div></div>
{{else}}

<div></div>
{{end}}`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: Replace,
			element:    dt.patches.actual.FirstChild.NextSibling,
			attr:       attrChange{},
		},
	}, t)
}

func TestDiff_DiffAttr(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<button {{if .Check}}disabled="disabled"{{end}}></button>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
	}, t)
}

func TestDiff_DiffAttrs(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<button {{if .Check}}disabled="disabled" class="hello world"{{end}}></button>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "class",
				value: "hello world",
			},
		},
	}, t)
}

func TestDiff_DiffMultiElementAndAttrs(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `<button {{if .Check}}disabled="disabled" class="hello world"{{end}}></button><button {{if .Check}}disabled="disabled" class="hello world"{{end}}></button>`,
	})

	dt.assert([]nodeExpect{
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild,
			attr: attrChange{
				name:  "class",
				value: "hello world",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild.NextSibling,
			attr: attrChange{
				name:  "disabled",
				value: "disabled",
			},
		},
		{
			changeType: SetAttr,
			element:    dt.patches.actual.FirstChild.NextSibling,
			attr: attrChange{
				name:  "class",
				value: "hello world",
			},
		},
	}, t)
}

func TestDiff_DiffMultiKey(t *testing.T) {
	t.Parallel()

	dt := newNodeTest(nodeTest{
		template: `
			<div key="1"></div>
			<div key="2"></div>
			{{ if not .Check }}
				<div key="3"></div>
			{{ end }}
			<div key="4">
				<b>Hello world</b>
			</div>
		`,
	})

	fmt.Println(dt.patches.updates)

}
