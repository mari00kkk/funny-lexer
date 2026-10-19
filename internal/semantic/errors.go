package semantic

type Error struct {
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
}

func (e Error) Error() string { return "semantic error: " + e.Message }
