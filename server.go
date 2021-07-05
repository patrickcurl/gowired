package gowired

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type WiredServer struct {
	// Wire ...
	Wire *LiveWire

	// CookieName ...
	CookieName string
	Log        Log
}

type WiredResponse struct {
	Rendered string
	Session  string
}

func NewServer() *WiredServer {
	logger := NewLoggerBasic()
	return &WiredServer{
		Wire:       NewWire(),
		CookieName: "_csrf_token",
		Log:        logger.Log,
	}
}

func (s *WiredServer) HandleFirstRequest(lc *WiredComponent, c PageContent) (*WiredResponse, error) {
	/* Create session to the new user */
	sessionKey, session, err := s.Wire.CreateSession()
	if err != nil {
		return nil, err
	}

	s.Log(LogInfo, "http request", logEx{"Component": lc.Name, "session": sessionKey})

	session.log = s.Log

	// Instantiate a page to attach to a session
	p := NewWiredPage(lc)

	// Set page content
	p.SetContent(c)

	// activation should be before mount,
	// because in activation will setup page channels
	// that will be needed in mount
	session.ActivatePage(p)

	// Mount page
	p.Mount()

	// Render page
	rendered, err := p.Render()

	if err != nil {
		return &WiredResponse{
			Rendered: "<h1> Page with error </h1>",
			Session:  "",
		}, err
	}

	return &WiredResponse{Rendered: rendered, Session: sessionKey}, nil
}

func (s *WiredServer) HandleHTMLRequest(ctx *fiber.Ctx, lc *WiredComponent, c PageContent) {

	lr, err := s.HandleFirstRequest(lc, c)

	if lr == nil {
		s.Log(LogPanic, "no wired page", logEx{"error": err})
		return
	}

	if err != nil {
		s.Log(LogError, "handle html request", logEx{"error": err})
		ctx.Response().SetStatusCode(500)
		return
	}

	ctx.Cookie(&fiber.Cookie{
		Name:    s.CookieName,
		Value:   lr.Session,
		Expires: time.Now().Add(24 * time.Hour),
	})

	ctx.Response().Header.SetContentType("text/html")
	ctx.Response().AppendBodyString(lr.Rendered)
}

func (s *WiredServer) CreateHTMLHandler(f func() *WiredComponent, c PageContent) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		lc := f()
		lc.log = s.Log

		s.HandleHTMLRequest(ctx, lc, c)
		return nil
	}
}

// HTTPMiddleware Middleware to run on HTTP requests.
type HTTPMiddleware func(next HTTPHandlerCtx) HTTPHandlerCtx

// HTTPHandlerCtx HTTP Handler with a page level context.
type HTTPHandlerCtx func(ctx *fiber.Ctx, pageCtx context.Context)

func (s *WiredServer) CreateHTMLHandlerWithMiddleware(f func(ctx context.Context) *WiredComponent, content PageContent,
	middlewares ...HTTPMiddleware) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		// TODO: move chain building out so it only happens once - Sam H.
		if len(middlewares) != 0 {
			// Reassign the context back to this scope to capture changes to it
			end := func(_ *fiber.Ctx, pCtx context.Context) {
				ctx = pCtx
			}
			h := middlewares[len(middlewares)-1](end)

			// build chain
			for i := len(middlewares) - 2; i >= 0; i-- {
				h = middlewares[i](h)
			}
			// trigger chain
			h(c, ctx)
		}

		lc := f(ctx)
		lc.log = s.Log

		s.HandleHTMLRequest(c, lc, content)

		return nil
	}
}

func (s *WiredServer) HandlePollRequest() {
	defer func() {
		payload := recover()
		if payload != nil {
			s.Log(LogWarn, fmt.Sprintf("ws request panic recovered: %v", payload), nil)
		}
	}()
}

func (s *WiredServer) HandleWSRequest(c *websocket.Conn) {
	defer func() {
		payload := recover()
		if payload != nil {
			s.Log(LogWarn, fmt.Sprintf("ws request panic recovered: %v", payload), nil)
		}
	}()

	c.EnableWriteCompression(true)

	sessionKey := c.Cookies(s.CookieName)

	s.Log(LogInfo, "websocket open", logEx{"session": sessionKey})

	session := s.Wire.GetSession(sessionKey)

	if session == nil || session.Status != SessionNew {
		s.Log(LogWarn, "session not found", logEx{"session": sessionKey})

		var msg PatchBrowser
		msg.Type = EventWiredError
		msg.Message = WiredErrorSessionNotFound
		if err := c.WriteJSON(msg); err != nil {
			s.Log(LogError, "handle ws request: write json", logEx{"error": err})
		}

		if err := c.Close(); err != nil {
			s.Log(LogError, "close websocket connection", logEx{"error": err})
		}

		s.Log(LogInfo, "websocket close", logEx{"session": sessionKey})

		return
	}

	session.Status = SessionOpen
	exit := make(chan int)
	exited := false

	go func() {
		for {
			select {
			case msg := <-session.OutChannel:
				s.Log(LogDebug, "message out", logEx{"msg": msg, "session": sessionKey})

				if err := c.WriteJSON(msg); err != nil {
					s.Log(LogError, "handle ws request: write json", logEx{"error": err})
				}
			case <-exit:
				exited = true

				session.Status = SessionClosed

				if err := c.Close(); err != nil {
					s.Log(LogError, "close websocket connection", logEx{"error": err})
				}

				if err := session.WiredPage.entryComponent.Kill(); err != nil {
					s.Log(LogError, "handle ws request: kill page", logEx{"error": err})
				}

				s.Wire.DeleteSession(sessionKey)

				s.Log(LogInfo, "websocket close", logEx{"session": sessionKey})

				return
			}
		}
	}()

	c.SetCloseHandler(func(code int, text string) error {
		// Close codes defined in RFC 6455, section 11.7.
		s.Log(LogTrace, "ws close handler", logEx{"code": code, "text": text})

		exit <- 1
		return nil
	})

	for {
		if exited {
			return
		}

		inMsg := BrowserEvent{}

		// Loop blocks here
		if err := c.ReadJSON(&inMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				// This seems to happen when running in Docker
				if !exited {
					s.Log(LogWarn, "handle ws request: unexpected connection close", nil)

					exit <- 1
				}

				return
			}

			s.Log(LogError, "handle ws request: read json", logEx{"error": err})

			continue
		}

		s.Log(LogDebug, "message in", logEx{"msg": inMsg, "session": sessionKey})

		if err := session.IngestMessage(inMsg); err != nil {
			s.Log(LogError, "handle ws request: ingest message", logEx{"error": err})
		}
	}
}
