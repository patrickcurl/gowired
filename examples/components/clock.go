package components

import (
	"time"

	"github.com/patrickcurl/gowired"
)

type Clock struct {
	gowired.LiveComponentWrapper
	ActualTime string
}

func NewClock() *gowired.LiveComponent {
	return gowired.NewLiveComponent("Clock", &Clock{
		ActualTime: formattedActualTime(),
	})
}

func formattedActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (c *Clock) Mounted(l *gowired.LiveComponent) {
	go func() {
		for {
			if l.Exited {
				return
			}
			c.ActualTime = formattedActualTime()
			time.Sleep(time.Second)
			c.Commit()
		}
	}()
}

func (c *Clock) TemplateHandler(_ *gowired.LiveComponent) string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}
