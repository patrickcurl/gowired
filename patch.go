package gowired

type Change struct {
	Name     string      `json:"n"`
	Type     string      `json:"t"`
	Attr     interface{} `json:"a,omitempty"`
	Content  string      `json:"c,omitempty"`
	Selector string      `json:"s"`
	Index    int         `json:"i,omitempty"`
}

type PatchNodeChildren map[int]*PatchTreeNode

type PatchTreeNode struct {
	Children    PatchNodeChildren  `json:"c,omitempty"`
	Changes []Change `json:"i"`
}

type PatchBrowser struct {
	ComponentID  string             `json:"cid,omitempty"`
	Type         string             `json:"t"`
	Message      string             `json:"m"`
	Changes []Change `json:"i,omitempty"`
}

func NewPatchBrowser(componentID string) *PatchBrowser {
	return &PatchBrowser{
		ComponentID:  componentID,
		Changes: make([]Change, 0),
	}
}

func (pb *PatchBrowser) appendChange(pi Change) {
	pb.Changes = append(pb.Changes, pi)
}
