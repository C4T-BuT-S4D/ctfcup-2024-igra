package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/samber/lo"

	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"
)

var interestingKeys = []ebiten.Key{
	ebiten.KeyA,
	ebiten.KeyB,
	ebiten.KeyC,
	ebiten.KeyD,
	ebiten.KeyE,
	ebiten.KeyF,
	ebiten.KeyG,
	ebiten.KeyH,
	ebiten.KeyI,
	ebiten.KeyJ,
	ebiten.KeyK,
	ebiten.KeyL,
	ebiten.KeyM,
	ebiten.KeyN,
	ebiten.KeyO,
	ebiten.KeyP,
	ebiten.KeyQ,
	ebiten.KeyR,
	ebiten.KeyS,
	ebiten.KeyT,
	ebiten.KeyU,
	ebiten.KeyV,
	ebiten.KeyW,
	ebiten.KeyX,
	ebiten.KeyY,
	ebiten.KeyZ,
	ebiten.KeyDigit0,
	ebiten.KeyDigit1,
	ebiten.KeyDigit2,
	ebiten.KeyDigit3,
	ebiten.KeyDigit4,
	ebiten.KeyDigit5,
	ebiten.KeyDigit6,
	ebiten.KeyDigit7,
	ebiten.KeyDigit8,
	ebiten.KeyDigit9,
	ebiten.KeySpace,
	ebiten.KeyComma,
	ebiten.KeyPeriod,
	ebiten.KeySlash,
	ebiten.KeyEscape,
	ebiten.KeyEnter,
	ebiten.KeyBackspace,
	ebiten.KeyShiftLeft,
	ebiten.KeyUp,
	ebiten.KeyDown,
}

type Input struct {
	pressedKeys      map[ebiten.Key]struct{}
	newlyPressedKeys map[ebiten.Key]struct{}
}

func New() *Input {
	return &Input{
		pressedKeys:      make(map[ebiten.Key]struct{}),
		newlyPressedKeys: make(map[ebiten.Key]struct{}),
	}
}

func NewFromProto(p *gameserverpb.ClientEvent_KeysPressed) *Input {
	i := New()
	for _, key := range p.KeysPressed {
		i.pressedKeys[ebiten.Key(key)] = struct{}{}
	}
	for _, key := range p.NewKeysPressed {
		i.newlyPressedKeys[ebiten.Key(key)] = struct{}{}
	}
	return i
}

func (i *Input) Update() {
	oldPressedKeys := i.pressedKeys

	i.pressedKeys = make(map[ebiten.Key]struct{})
	i.newlyPressedKeys = make(map[ebiten.Key]struct{})
	for _, key := range interestingKeys {
		if ebiten.IsKeyPressed(key) {
			i.pressedKeys[key] = struct{}{}
			if _, ok := oldPressedKeys[key]; !ok {
				i.newlyPressedKeys[key] = struct{}{}
			}
		}
	}
}

func (i *Input) IsKeyPressed(key ebiten.Key) bool {
	_, ok := i.pressedKeys[key]
	return ok
}

func (i *Input) IsKeyNewlyPressed(key ebiten.Key) bool {
	_, ok := i.newlyPressedKeys[key]
	return ok
}

func (i *Input) JustPressedKeys() []ebiten.Key {
	return lo.Uniq(lo.Keys(i.newlyPressedKeys))
}

func (i *Input) ToProto() *gameserverpb.ClientEvent_KeysPressed {
	return &gameserverpb.ClientEvent_KeysPressed{
		KeysPressed: lo.Map(lo.Keys(i.pressedKeys), func(key ebiten.Key, _ int) int32 {
			return int32(key)
		}),
		NewKeysPressed: lo.Map(lo.Keys(i.newlyPressedKeys), func(key ebiten.Key, _ int) int32 {
			return int32(key)
		}),
	}
}
