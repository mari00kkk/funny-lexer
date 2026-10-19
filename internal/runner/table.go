package runner

import (
	"encoding/json"
	"os"

	"funny-lexer/internal/dfa"
)

type TransitionDTO struct {
	From int   `json:"from"`
	Char int32 `json:"char"`
	To   int   `json:"to"`
}

type DFATableDTO struct {
	StartState   int            `json:"start_state"`
	AcceptStates map[int]string `json:"accept_states"`
	Transitions  []TransitionDTO `json:"transitions"`
}

// ExportTableJSON сериализует минимизированный ДКА в JSON
func ExportTableJSON(d *dfa.DFA, filepath string) error {
	dto := DFATableDTO{
		StartState:   d.Start.ID,
		AcceptStates: make(map[int]string),
		Transitions:  make([]TransitionDTO, 0),
	}

	for _, state := range d.States {
		if state.IsAccept {
			dto.AcceptStates[state.ID] = state.Token
		}
		for ch, next := range state.Transitions {
			if next != nil {
				dto.Transitions = append(dto.Transitions, TransitionDTO{
					From: state.ID,
					Char: ch,
					To:   next.ID,
				})
			}
		}
	}

	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}