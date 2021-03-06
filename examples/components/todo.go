package components

import (
	"strings"

	"github.com/patrickcurl/gowired"
)

type Task struct {
	Done bool
	Text string
}

func (t *Task) GetClasses() string {
	classes := []string{
		"task",
	}

	if t.Done {
		classes = append(classes, "active")
	}

	return strings.Join(classes, " ")
}

type Todo struct {
	gowired.WiredComponentWrapper
	Counter int
	Text    string
	Tasks   []Task
	Name    string
}

func NewTodo() *gowired.WiredComponent {
	return gowired.NewWiredComponent("Todo", &Todo{
		Counter: 0,
		Name:    "Todo",
		Text:    "",
		Tasks: []Task{
			{
				Done: true,
				Text: "Wake up",
			},
			{
				Done: true,
				Text: "Breath",
			},
			{
				Done: false,
				Text: "Turn on the coffee maker",
			},
		},
	})
}

func (t *Todo) IncreaseCounter() {
	t.Counter++
}

func (t *Todo) HandleAdd() {
	if len(t.Text) > 0 {
		t.Tasks = append(t.Tasks, Task{
			Done: false,
			Text: t.Text,
		})
		t.Text = ""
	}
}

func (t *Todo) TaskDone(index int) {
	t.Tasks[index].Done = true
}

func (t *Todo) CanAdd() bool {
	return len(t.Text) > 0
}

func (t *Todo) TemplateHandler(_ *gowired.WiredComponent) string {
	return `
		<div id="todo">
			<input go-wired-input="Text" />
			<button :disabled="{{not .CanAdd}}" go-wired-click="HandleAdd">Create</button>
			<div class="todo-tasks">
				{{ range $index, $task := .Tasks }}
					<div class="{{ $task.GetClasses }}" key="{{$index}}">
						<input type="checkbox" go-wired-input="Tasks.{{$index}}.Done"></input>
						<span>{{ $task.Text }}</span>
					</div>
				{{ end }}
			</div>

			<style>
				.task {
					padding: 10px 20px;
					rounded: 20px;
				}
				.active {
				    color: rgba(0,0,0,0.2);
    				text-decoration: line-through;
				}
			</style>
		</div>
	`
}
