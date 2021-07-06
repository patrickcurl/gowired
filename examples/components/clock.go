package components

import (
	"time"

	"github.com/patrickcurl/gowired"
)

type Clock struct {
	gowired.WiredComponentWrapper
	ActualTime string
}

func NewClock() *gowired.WiredComponent {
	return gowired.NewWiredComponent("Clock", &Clock{
		ActualTime: formattedActualTime(),
	})
}

func formattedActualTime() string {
	return time.Now().Format(time.RFC3339Nano)
}

func (c *Clock) Mounted(l *gowired.WiredComponent) {
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

func (c *Clock) TemplateHandler(_ *gowired.WiredComponent) string {
	return `
		<div>
			<span>Time: {{ .ActualTime }}</span>
		</div>
	`
}
