package gowired

import (
	"bytes"
	"fmt"
	"html/template"

	"golang.org/x/net/html"
)

type WiredState struct {
	html *html.Node
	text string
}

func (state *WiredState) setText(text string) error {
	var err error
	state.html, err = nodeFromString(text)
	state.text = text
	return err
}

func (state *WiredState) setHTML(node *html.Node) error {
	var err error
	state.text, err = renderInnerHTML(node)
	state.html = node
	return err
}

type WiredRenderer struct {
	state          *WiredState
	template       *template.Template
	templateString string
	formatters     []func(t string) string
}

func (renderer *WiredRenderer) setTemplate(t *template.Template, ts string) {
	renderer.template = t
	renderer.templateString = ts
}

func (renderer *WiredRenderer) renderToText(data interface{}) (string, error) {
	if renderer.template == nil {
		return "", fmt.Errorf("template is not defined in WiredRenderer")
	}

	s := bytes.NewBufferString("")

	err := renderer.template.Execute(s, data)

	if err != nil {
		err = fmt.Errorf("template execute: %w", err)
	}

	text := s.String()
	for _, f := range renderer.formatters {
		text = f(text)
	}

	return text, err
}

func (renderer *WiredRenderer) Render(data interface{}) (string, *html.Node, error) {

	textRender, err := renderer.renderToText(data)
	if err != nil {
		return "", nil, err
	}

	err = renderer.state.setText(textRender)
	return renderer.state.text, renderer.state.html, err
}

func (renderer *WiredRenderer) WiredRender(data interface{}) (*patches, error) {

	actualRender := renderer.state.html
	proposedRenderText, err := renderer.renderToText(data)

	err = renderer.state.setText(proposedRenderText)

	if err != nil {
		return nil, fmt.Errorf("state set text: %w", err)
	}

	// TODO: maybe the right way to call a diff is calling based on state
	changed := updateNode(actualRender)
	changed.propose(renderer.state.html)

	return changed, nil
}

func (renderer *WiredRenderer) useFormatter(f func(t string) string) {
	renderer.formatters = append(renderer.formatters, f)
}
