package metadata

import "fmt"

type ModThunk struct {
	Id      int
	Name    string
	Content string
}

func NewModThunk(id int, name, content string) ModThunk {
	return ModThunk{
		id,
		name,
		content,
	}
}

func (m ModThunk) ToNameString() string {
	return fmt.Sprintf("%v_%v", m.Id, m.Name)
}
