package gowired

import (
	"fmt"
	"strings"
)

const (
	EventWiredInput          = "li"
	EventWiredMethod         = "lm"
	EventWiredDom            = "ld"
	EventWiredDisconnect     = "lx"
	EventWiredError          = "le"
	EventWiredConnectElement = "lce"
)

var (
	WiredErrorSessionNotFound = "session_not_found"
)

func WiredErrorMap() map[string]string {
	return map[string]string{
		"WiredErrorSessionNotFound": WiredErrorSessionNotFound,
	}
}

type BrowserEvent struct {
	Name        string            `json:"name"`
	ComponentID string            `json:"component_id"`
	MethodName  string            `json:"method_name"`
	MethodData  map[string]string `json:"method_data"`
	StateKey    string            `json:"key"`
	StateValue  string            `json:"value"`
	DOMEvent    *DOMEvent         `json:"dom_event"`
}

type DOMEvent struct {
	KeyCode string `json:"keyCode"`
}

type SessionStatus string

const (
	SessionNew    SessionStatus = "n"
	SessionOpen   SessionStatus = "o"
	SessionClosed SessionStatus = "c"
)

type Session struct {
	WiredPage   *Page
	OutChannel chan PatchBrowser
	log        Log
	Status     SessionStatus
}

func NewSession() *Session {
	return &Session{
		OutChannel: make(chan PatchBrowser),
		Status:     SessionNew,
	}
}

func (session *Session) QueueMessage(message PatchBrowser) {
	go func() {
		session.OutChannel <- message
	}()
}

func (session *Session) IngestMessage(message BrowserEvent) error {

	defer func() {
		payload := recover()
		if payload != nil {
			// TODO: get session key in log
			session.log(LogWarn, fmt.Sprintf("ingest message: recover from panic: %v", payload), logEx{"message": message})
		}
	}()

	err := session.WiredPage.HandleBrowserEvent(message)

	if err != nil {
		return err
	}

	return nil
}

func (session *Session) ActivatePage(lp *Page) {
	session.WiredPage = lp

	// Here is the location that get all the components updates *notified* by
	// the page!
	go func() {
		for {
			// Receive all the events from page
			evt := <-session.WiredPage.Events

			session.log(LogDebug, fmt.Sprintf("Component %s triggering %d", evt.Component.Name, evt.Type), logEx{"evt": evt})

			switch evt.Type {
			case PageComponentUpdated:
				if err := session.WiredRenderComponent(evt.Component, evt.Source); err != nil {
					session.log(LogError, "entryComponent wired render", logEx{"error": err})
				}
				break
			case PageComponentMounted:
				session.QueueMessage(PatchBrowser{
					ComponentID:  evt.Component.Name,
					Type:         EventWiredConnectElement,
					Instructions: nil,
				})
				break
			}
		}
	}()
}

func (session *Session) generateBrowserPatchesFromDiff(diff *diff, source *EventSource) ([]*PatchBrowser, error) {

	patchedBrowser := make([]*PatchBrowser, 0)

	for _, patched := range diff.patches {

		selector, err := selectorFromNode(patched.element)
		if skipUpdateValueOnInput(patched, source) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("selector from node: %w patched: %v", err, patched)
		}

		componentID, err := componentIDFromNode(patched.element)

		if err != nil {
			return nil, err
		}

		var patch *PatchBrowser

		// find if there is already a patch
		for _, patchedBrowser := range patchedBrowser {
			if patchedBrowser.ComponentID == componentID {
				patch = patchedBrowser
				break
			}
		}

		// If there is no patch
		if patch == nil {
			patch = NewPatchBrowser(componentID)
			patch.Type = EventWiredDom
			patchedBrowser = append(patchedBrowser, patch)
		}

		patch.patchNode(PatchedNode{
			Name: EventWiredDom,
			Type: patched.changeType.toString(),
			Attr: map[string]string{
				"Name":  patched.attr.name,
				"Value": patched.attr.value,
			},
			Index:    patched.index,
			Content:  patched.content,
			Selector: selector.toString(),
		})
	}
	return patchedBrowser, nil
}

func skipUpdateValueOnInput(updated updateInstruction, source *EventSource) bool {
	if updated.element == nil || source == nil || updated.updateType != SetAttr || strings.ToLower(updated.attr.name) != "value" {
		return false
	}

	attr := getAttribute(updated.element, "go-wired-input")

	return attr != nil && source.Type == EventSourceInput && attr.Val == source.Value
}

// WiredRenderComponent render the updated Component and compare with
// last state. It may apply with *all child components*
func (s *Session) WiredRenderComponent(c *WiredComponent, source *EventSource) error {
	var err error

	diff, err := c.WiredRender()

	if err != nil {
		return err
	}

	patches, err := s.generateBrowserPatchesFromDiff(diff, source)

	if err != nil {
		return err
	}

	for _, om := range patches {
		s.QueueMessage(*om)
	}

	return nil
}
