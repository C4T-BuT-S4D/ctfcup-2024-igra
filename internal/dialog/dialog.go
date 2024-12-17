package dialog

import gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"

type State struct {
	Finished bool
	GaveItem bool
	Text     string
}

func (s *State) ToProto() *gameserverpb.DialogState {
	return &gameserverpb.DialogState{
		Finished: s.Finished,
		GaveItem: s.GaveItem,
		Text:     s.Text,
	}
}

func StateFromProto(s *gameserverpb.DialogState) *State {
	return &State{
		Finished: s.Finished,
		GaveItem: s.GaveItem,
		Text:     s.Text,
	}
}

type Dialog interface {
	Greeting()
	Feed(text string, coins int)
	State() *State
	SetState(s *State)
}
