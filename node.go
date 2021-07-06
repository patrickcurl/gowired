package gowired

import (
	"strconv"

	"golang.org/x/net/html"
)

type ChangeType int

func (ct ChangeType) toString() string {
	return strconv.Itoa(int(ct))
}

const (
	Append ChangeType = iota
	Remove
	SetInnerHTML
	SetAttr
	RemoveAttr
	Replace
	Move
)

type node struct {
	changeType ChangeType
	element    *html.Node
	content    string
	attr       attrChange
	index      int
}

// attrChange todo
type attrChange struct {
	name  string
	value string
}

type patches struct {
	actual       *html.Node
	updates []node
	quantity     int
	doneElements []*html.Node
}

func updateNode(actual *html.Node) *patches {
	return &patches{
		actual:       actual,
		updates: make([]node, 0),
	}
}

func (p *patches) updatesByType(t ChangeType) []node {
	s := make([]node, 0)

	for _, i := range p.updates {
		if i.changeType == t {
			s = append(s, i)
		}
	}

	return s
}

func (p *patches) checkpoint() {
	p.quantity = len(p.updates)
}

// Has changed since last checkpoint
func (p *patches) hasChanged() bool {
	return len(p.updates) != p.quantity
}

func (p *patches) propose(proposed *html.Node) {
	p.clearMarked()
	p.updateNode(p.actual, proposed)
}

func (p *patches) updateNode(actual, proposed *html.Node) {

	uidActual, actualOk := getWiredUidAttributeValue(actual)
	uidProposed, proposedOk := getWiredUidAttributeValue(proposed)

	if actualOk && proposedOk && uidActual != uidProposed {
		content, _ := renderNodeToString(proposed)
		p.updates = append(p.updates, node{
			changeType: Replace,
			element:    actual,
			content:    content,
		})
		return
	}

	p.updateNodeAttributes(actual, proposed)
	p.nodeWalk(actual.FirstChild, proposed.FirstChild)
	p.markNodeDone(proposed)
}

func (p *patches) clearMarked() {
	p.doneElements = make([]*html.Node, 0)
}

func (p *patches) markNodeDone(node *html.Node) {
	p.doneElements = append(p.doneElements, node)
}

func (p *patches) isMarked(node *html.Node) bool {
	for _, n := range p.doneElements {
		if n == node {
			return true
		}
	}

	return false
}

func (p *patches) nodeWalk(actual, proposed *html.Node) {

	if actual == nil && proposed == nil {
		return
	}

	if nodeIsText(actual) || nodeIsText(proposed) {
		p.checkpoint()
		p.updateTextNode(actual, proposed)
		if p.hasChanged() {
			return
		}
	}

	if actual != nil && proposed != nil {
		p.updateNode(actual, proposed)
	} else if actual == nil && nodeIsElement(proposed) {
		nodeContent, _ := renderNodeToString(proposed)
		p.updates = append(p.updates, node{
			changeType: Append,
			element:    proposed.Parent,
			content:    nodeContent,
		})
		p.markNodeDone(proposed)
	} else if proposed == nil && nodeIsElement(actual) {
		p.updates = append(p.updates, node{
			changeType: Remove,
			element:    actual,
		})
		p.markNodeDone(actual)
	}

	nextActual := nextRelevantElement(actual)
	nextProposed := nextRelevantElement(proposed)

	if nextActual != nil || nextProposed != nil {
		p.nodeWalk(nextActual, nextProposed)
	}
}

func (p *patches) forceRenderElementContent(proposed *html.Node) {
	childrenHTML, _ := renderInnerHTML(proposed)

	p.updates = append(p.updates, node{
		changeType: SetInnerHTML,
		content:    childrenHTML,
		element:    proposed,
	})
}

// patchNodeAttributes compares the attributes in el to the attributes in otherEl
// and adds the necessary patches to make the attributes in el match those in
// otherEl
func (p *patches) updateNodeAttributes(actual, proposed *html.Node) {

	actualAttrs := AttrMapFromNode(actual)
	proposedAttrs := AttrMapFromNode(proposed)

	// Now iterate through the attributes in otherEl
	for name, otherValue := range proposedAttrs {
		value, found := actualAttrs[name]
		if !found || value != otherValue {
			p.updates = append(p.updates, node{
				changeType: SetAttr,
				element:    actual,
				attr: attrChange{
					name:  name,
					value: otherValue,
				},
			})
		}
	}

	for attrName := range actualAttrs {
		if _, found := proposedAttrs[attrName]; !found {

			p.updates = append(p.updates, node{
				changeType: RemoveAttr,
				element:    actual,
				attr: attrChange{
					name: attrName,
				},
			})
		}
	}
}

func (p *patches) updateTextNode(actual, proposed *html.Node) {

	// Any node is text
	if !nodeIsText(proposed) && !nodeIsText(actual) {
		return
	}

	proposedIsRelevant := nodeRelevant(proposed)
	actualIsRelevant := nodeRelevant(actual)

	if !proposedIsRelevant && !actualIsRelevant {
		return
	}

	// XOR
	if proposedIsRelevant != actualIsRelevant {
		goto renderEntireNode
	}

	if proposed.Data != actual.Data {
		goto renderEntireNode
	}

	return

renderEntireNode:
	{

		node := proposed

		if node == nil {
			node = actual
		}

		if node == nil {
			return
		}

		p.forceRenderElementContent(node.Parent)
		p.markNodeDone(node.Parent)
	}

}
