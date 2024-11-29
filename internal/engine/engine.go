package engine

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Rulox/ebitmx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/salviati/go-tmx/tmx"
	"github.com/samber/lo"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
	"golang.org/x/image/font"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/dialog"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/fonts"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/input"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/music"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/npc"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/physics"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/player"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/portal"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/sprites"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/tiles"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/wall"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"

	// Register png codec.
	_ "image/png"
)

const dialogShowLines = 12

type Factory func() (*Engine, error)

type Config struct {
	SnapshotsDir string
	Level        string
}

type dialogControl struct {
	inputBuffer []rune
	scroll      int
	maskInput   bool
}

type Engine struct {
	Tiles        []*tiles.StaticTile `json:"-" msgpack:"-"`
	Camera       *camera.Camera      `json:"-" msgpack:"camera"`
	Player       *player.Player      `json:"-" msgpack:"player"`
	Items        []*item.Item        `json:"items" msgpack:"items"`
	Portals      []*portal.Portal    `json:"-" msgpack:"portals"`
	Spikes       []*damage.Spike     `json:"-" msgpack:"spikes"`
	InvWalls     []*wall.InvWall     `json:"-" msgpack:"invWalls"`
	NPCs         []*npc.NPC          `json:"-" msgpack:"npcs"`
	EnemyBullets []*damage.Bullet    `json:"-" msgpack:"enemyBullets"`

	StartSnapshot *Snapshot `json:"-" msgpack:"-"`

	fontsManager  *fonts.Manager
	spriteManager *sprites.Manager
	musicManager  *music.Manager
	snapshotsDir  string
	playerSpawn   *geometry.Point
	activeNPC     *npc.NPC
	dialogControl dialogControl

	Muted    bool   `json:"-" msgpack:"-"`
	Paused   bool   `json:"-" msgpack:"paused"`
	Tick     int    `json:"-" msgpack:"tick"`
	Level    string `json:"-" msgpack:"level"`
	IsWin    bool   `json:"-" msgpack:"isWin"`
	TeamName string `json:"-" msgpack:"-"`
}

func getProperties(o *tmx.Object) map[string]string {
	properties := make(map[string]string)
	for _, p := range o.Properties {
		properties[p.Name] = p.Value
	}
	return properties
}

var ErrNoPlayerSpawn = errors.New("no player spawn found")

func findPlayerSpawn(tileMap *tmx.Map) (*geometry.Point, error) {
	for _, og := range tileMap.ObjectGroups {
		for _, o := range og.Objects {
			if o.Type == "player_spawn" {
				return &geometry.Point{
					X: o.X,
					Y: o.Y,
				}, nil
			}
		}
	}

	return nil, ErrNoPlayerSpawn
}

func getTileImgByID(tileID tmx.ID, tileSet *ebitmx.EbitenTileset, img *ebiten.Image) *ebiten.Image {
	id := int(tileID)

	x0 := (id % tileSet.TilesetWidth) * tileSet.TileWidth
	y0 := (id / tileSet.TilesetWidth) * tileSet.TileHeight
	x1, y1 := x0+tileSet.TileWidth, y0+tileSet.TileHeight

	return img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
}

func New(config Config, spriteManager *sprites.Manager, fontsManager *fonts.Manager, musicManager *music.Manager, dialogProvider dialog.Provider) (*Engine, error) {
	var resultImage *ebiten.Image
	imgFile, err := resources.EmbeddedFS.Open("tiles/result.png")
	if err != nil {
		return nil, fmt.Errorf("failed to open tileset: %w", err)
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tileset: %w", err)
	}

	resultImage = ebiten.NewImageFromImage(img)

	mapFile, err := resources.EmbeddedFS.Open(fmt.Sprintf("levels/%s.tmx", config.Level))
	if err != nil {
		return nil, fmt.Errorf("failed to open map: %w", err)
	}

	testMap, err := tmx.Read(mapFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map: %w", err)
	}

	iceTiles, err := ebitmx.GetTilesetFromFS(resources.EmbeddedFS, "tiles/ice.tsx")
	if err != nil {
		return nil, fmt.Errorf("failed to get tileset: %w", err)
	}

	var mapTiles []*tiles.StaticTile

	for _, l := range testMap.Layers {
		for x := 0; x < testMap.Width; x++ {
			for y := 0; y < testMap.Height; y++ {
				dt := l.DecodedTiles[y*testMap.Width+x]
				if dt.IsNil() {
					continue
				}

				mapTiles = append(
					mapTiles,
					tiles.NewStaticTile(
						&geometry.Point{
							X: float64(x * testMap.TileWidth),
							Y: float64(y * testMap.TileHeight),
						},
						testMap.TileWidth,
						testMap.TileHeight,
						getTileImgByID(dt.ID, iceTiles, resultImage),
					),
				)
			}
		}
	}

	playerPos, err := findPlayerSpawn(testMap)
	if err != nil {
		return nil, fmt.Errorf("can't find player position: %w", err)
	}

	p, err := player.New(playerPos, spriteManager)
	if err != nil {
		return nil, fmt.Errorf("creating player: %w", err)
	}

	var items []*item.Item
	var spikes []*damage.Spike
	var invwalls []*wall.InvWall
	var npcs []*npc.NPC
	winPoints := make(map[string]*geometry.Point)
	portalsMap := make(map[string]*portal.Portal)

	for _, og := range testMap.ObjectGroups {
		for _, o := range og.Objects {
			props := getProperties(&o)
			switch o.Type {
			case "item":
				img := ebiten.NewImage(int(o.Width), int(o.Height))
				img.Fill(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff})

				if sprite := props["sprite"]; sprite != "" {
					img = spriteManager.GetSprite(sprites.Type(sprite))
				}

				items = append(items, item.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					o.Width,
					o.Height,
					img,
					o.Name,
					props["important"] == "true",
				))
			case "portal":
				img := spriteManager.GetSprite(sprites.Portal)
				portalsMap[o.Name] = portal.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					o.Width,
					o.Height,
					props["portal-to"],
					nil,
					props["boss"])
			case "spike":
				img := spriteManager.GetSprite(sprites.Spike)

				spikes = append(spikes, damage.NewSpike(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					o.Width,
					o.Height,
				))
			case "invwall":
				invwalls = append(invwalls, wall.NewInvWall(&geometry.Point{
					X: o.X,
					Y: o.Y,
				},
					o.Width,
					o.Height))
			case "npc":
				img := spriteManager.GetSprite(sprites.Type(props["sprite"]))
				dimg := spriteManager.GetSprite(sprites.Type(props["dialog-sprite"]))
				npcd, err := dialogProvider.Get(o.Name)
				if err != nil {
					return nil, fmt.Errorf("getting '%s' dialog: %w", o.Name, err)
				}
				npcs = append(npcs, npc.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					dimg,
					o.Width,
					o.Height,
					npcd,
					props["item"],
				))
			case "boss-win":
				winPoints[o.Name] = &geometry.Point{X: o.X, Y: o.Y}
			}
		}
	}

	for _, n := range npcs {
		_, i, ok := lo.FindIndexOf(items, func(i *item.Item) bool {
			return i.Name == n.ReturnsItem
		})
		if !ok {
			return nil, fmt.Errorf("item %s not found for npc", n.ReturnsItem)
		}
		n.LinkedItem = items[i]
	}

	for name, p := range portalsMap {
		if p.PortalTo == "" {
			continue
		}
		toPortal := portalsMap[p.PortalTo]
		if toPortal == nil {
			return nil, fmt.Errorf("destination %s not found for portal %s", p.PortalTo, name)
		}
		p.TeleportTo = toPortal.Origin.Add(&geometry.Vector{
			X: 32,
			Y: 0,
		})
	}

	cam := &camera.Camera{
		Object: &object.Object{
			Origin: &geometry.Point{
				X: 0,
				Y: 0,
			},
			Width:  camera.WIDTH,
			Height: camera.HEIGHT,
		},
	}

	keys := lo.Keys(portalsMap)
	slices.Sort(keys)
	portals := make([]*portal.Portal, 0, len(keys))
	for _, key := range keys {
		portals = append(portals, portalsMap[key])
	}

	return &Engine{
		Tiles:         mapTiles,
		Camera:        cam,
		Player:        p,
		Items:         items,
		Portals:       portals,
		Spikes:        spikes,
		InvWalls:      invwalls,
		NPCs:          npcs,
		spriteManager: spriteManager,
		fontsManager:  fontsManager,
		musicManager:  musicManager,
		snapshotsDir:  config.SnapshotsDir,
		playerSpawn:   playerPos,
		Level:         config.Level,
		TeamName:      strings.Split(os.Getenv("AUTH_TOKEN"), ":")[0],
		dialogControl: dialogControl{
			maskInput: !dialogProvider.DisplayInput(),
		},
	}, nil
}

func NewFromSnapshot(config Config, snapshot *Snapshot, spritesManager *sprites.Manager, fontsManager *fonts.Manager, musicManager *music.Manager, dialogProvider dialog.Provider) (*Engine, error) {
	e, err := New(config, spritesManager, fontsManager, musicManager, dialogProvider)
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}

	e.StartSnapshot = snapshot

	if err := json.Unmarshal(snapshot.Data, e); err != nil {
		return nil, fmt.Errorf("applying snapshot: %w", err)
	}

	for _, it := range e.Items {
		if it.Collected {
			e.Player.Inventory.Items = append(e.Player.Inventory.Items, it)
		}
	}

	return e, nil
}

type Snapshot struct {
	Data []byte
}

func NewSnapshotFromProto(proto *gameserverpb.EngineSnapshot) *Snapshot {
	return &Snapshot{Data: proto.Data}
}

func (s *Snapshot) ToProto() *gameserverpb.EngineSnapshot {
	if s == nil {
		return nil
	}
	return &gameserverpb.EngineSnapshot{
		Data: s.Data,
	}
}

func (e *Engine) Reset() {
	e.Player.MoveTo(e.playerSpawn)
	e.Player.Health = player.DefaultHealth
	e.activeNPC = nil
	e.EnemyBullets = nil
	e.Tick = 0
}

func (e *Engine) MakeSnapshot() (*Snapshot, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshalling engine: %w", err)
	}

	return &Snapshot{
		Data: data,
	}, nil
}

func (e *Engine) SaveSnapshot(snapshot *Snapshot) error {
	if e.snapshotsDir == "" {
		return nil
	}

	filename := fmt.Sprintf("snapshot_%s_%s", e.Level, time.Now().UTC().Format("2006-01-02T15:04:05.999999999"))

	if err := os.WriteFile(filepath.Join(e.snapshotsDir, filename), snapshot.Data, 0o400); err != nil {
		return fmt.Errorf("writing snapshot file: %w", err)
	}

	return nil
}

func (e *Engine) drawDiedScreen(screen *ebiten.Image) {
	face := e.fontsManager.Get(fonts.DSouls)
	redColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	width := font.MeasureString(face, "YOU DIED")

	text.Draw(screen, "YOU DIED", face, camera.WIDTH/2-width.Floor()/2, camera.HEIGHT/2, redColor)
}

func (e *Engine) drawYouWinScreen(screen *ebiten.Image) {
	face := e.fontsManager.Get(fonts.DSouls)
	gColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	width := font.MeasureString(face, "YOU WIN")

	text.Draw(screen, "YOU WIN", face, camera.WIDTH/2-width.Floor()/2, camera.HEIGHT/2, gColor)
}

func (e *Engine) drawNPCDialog(screen *ebiten.Image) {
	colorWhite := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	// Draw dialog border (outer rectangle).
	borderw, borderh := camera.WIDTH-camera.WIDTH/8, camera.HEIGHT/2
	img := ebiten.NewImage(borderw, borderh)
	img.Fill(colorWhite)
	op := &ebiten.DrawImageOptions{}
	bx, by := camera.WIDTH/16.0, camera.HEIGHT/2.0-camera.HEIGHT/16
	op.GeoM.Translate(bx, by)
	screen.DrawImage(img, op)

	// Draw dialog border (inner rectangle).
	ibw, ibh := borderw-camera.WIDTH/32, borderh-camera.HEIGHT/32
	ibx, iby := bx+camera.WIDTH/64, by+camera.HEIGHT/64
	img = ebiten.NewImage(ibw, ibh)
	img.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0xff})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ibx, iby)
	screen.DrawImage(img, op)

	// Draw dialog NPC image.
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(camera.WIDTH/2+camera.WIDTH/8, camera.HEIGHT/2)
	screen.DrawImage(e.activeNPC.DialogImage, op)

	// Draw dialog text.
	dtx, dty := ibx+camera.WIDTH/32, iby+camera.HEIGHT/32
	face := e.fontsManager.Get(fonts.Dialog)
	txt := e.activeNPC.Dialog.State().Text

	lines := input.AutoWrap(txt, face, ibw-camera.WIDTH/32)
	e.dialogControl.scroll = max(min(e.dialogControl.scroll, len(lines)-1), 0)

	l := e.dialogControl.scroll
	r := min(e.dialogControl.scroll+dialogShowLines, len(lines))

	visibleLines := lines[l:r]
	text.Draw(screen, strings.Join(visibleLines, "\n"), face, int(dtx), int(dty), colorWhite)

	// Draw dialog input buffer.
	if len(e.dialogControl.inputBuffer) > 0 {
		dtbx, dtby := dtx, dty+float64((len(visibleLines)-1)*face.Metrics().Height.Floor())+1.0*float64(face.Metrics().Height.Floor())
		c := color.RGBA{R: 0x00, G: 0xff, B: 0xff, A: 0xff}
		ibuf := string(e.dialogControl.inputBuffer)
		if e.dialogControl.maskInput {
			ibuf = strings.Repeat("*", len(ibuf))
		}
		x := input.AutoWrap(ibuf, face, ibw-camera.WIDTH/32)
		text.Draw(screen, strings.Join(x, "\n"), face, int(dtbx), int(dtby), c)
	}
}

func (e *Engine) Draw(screen *ebiten.Image) {
	if e.Player.IsDead() {
		e.drawDiedScreen(screen)
		return
	}

	if e.IsWin {
		e.drawYouWinScreen(screen)
		return
	}

	for _, c := range e.Collisions(e.Camera.Rectangle()) {
		visible := c.Rectangle().Sub(e.Camera.Rectangle())
		base := geometry.Origin.Add(visible)
		op := &ebiten.DrawImageOptions{}

		switch c.Type() {
		case object.PlayerType:
			if e.Player.LooksRight {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(e.Player.Width, 0)
			}
		case object.EnemyBullet:
			op.GeoM.Scale(4, 4)
			op.GeoM.Translate(-2, 0)
		default:
			// not a player or boss.
		}

		op.GeoM.Translate(
			base.X,
			base.Y,
		)

		switch c.Type() {
		case object.StaticTileType:
			t := c.(*tiles.StaticTile)
			screen.DrawImage(t.Image, op)
		case object.Item:
			it := c.(*item.Item)
			if it.Collected {
				continue
			}
			screen.DrawImage(it.Image, op)
		case object.PlayerType:
			screen.DrawImage(e.Player.Image(), op)
		case object.Portal:
			p := c.(*portal.Portal)
			screen.DrawImage(p.Image, op)
		case object.Spike:
			d := c.(*damage.Spike)
			screen.DrawImage(d.Image, op)
		case object.NPC:
			n := c.(*npc.NPC)
			screen.DrawImage(n.Image, op)
		case object.EnemyBullet:
			b := c.(*damage.Bullet)
			if !b.Triggered {
				screen.DrawImage(b.Image, op)
			}
		default:
		}
	}

	if !e.Player.IsDead() {
		face := e.fontsManager.Get(fonts.Dialog)

		teamtxt := fmt.Sprintf("Team %s", e.TeamName)
		text.Draw(screen, teamtxt, face, 72, 72, color.RGBA{R: 204, G: 14, B: 206, A: 255})

		txt := fmt.Sprintf("HP: %d", e.Player.Health)
		text.Draw(screen, txt, face, 72, 72+36, color.RGBA{R: 0, G: 255, B: 0, A: 255})

		for i, it := range e.Player.Inventory.Items {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(e.Camera.Width-float64(i+1)*72, 72)
			screen.DrawImage(it.Image, op)
		}
	}

	if e.activeNPC != nil {
		e.drawNPCDialog(screen)
	}
}

func (e *Engine) Update(inp *input.Input) error {
	e.Tick++

	if e.musicManager != nil {
		p := e.musicManager.GetPlayer(music.Background)
		if !e.Muted {
			p.Play()
		}
		if !e.Muted && !p.IsPlaying() {
			if err := p.Rewind(); err != nil {
				panic(err)
			}
		}
		if inp.IsKeyNewlyPressed(ebiten.KeyM) {
			e.Muted = !e.Muted
			if e.Muted {
				p.Pause()
			}
		}
	}

	if e.activeNPC != nil {
		if inp.IsKeyNewlyPressed(ebiten.KeyEscape) {
			e.activeNPC = nil
			e.dialogControl.inputBuffer = e.dialogControl.inputBuffer[:0]
			return nil
		}
		if e.activeNPC.Dialog.State().GaveItem {
			e.activeNPC.LinkedItem.MoveTo(e.activeNPC.Origin.Add(&geometry.Vector{
				X: +64,
				Y: +32,
			}))
		}

		pk := inp.JustPressedKeys()
		if len(pk) > 0 && !e.activeNPC.Dialog.State().Finished {
			c := pk[0]
			switch c {
			case ebiten.KeyUp:
				// TODO(scroll up)
				e.dialogControl.scroll--
			case ebiten.KeyDown:
				e.dialogControl.scroll++
			case ebiten.KeyBackspace:
				// backspace
				if len(e.dialogControl.inputBuffer) > 0 {
					e.dialogControl.inputBuffer = e.dialogControl.inputBuffer[:len(e.dialogControl.inputBuffer)-1]
				}
			case ebiten.KeyEnter:
				// enter
				e.activeNPC.Dialog.Feed(string(e.dialogControl.inputBuffer))
				e.dialogControl.inputBuffer = e.dialogControl.inputBuffer[:0]
			default:
				e.dialogControl.inputBuffer = append(e.dialogControl.inputBuffer, input.Key(c).Rune())
			}
		}

		return nil
	}

	if e.Paused {
		if inp.IsKeyNewlyPressed(ebiten.KeyP) {
			e.Paused = false
		} else {
			return nil
		}
	} else if inp.IsKeyNewlyPressed(ebiten.KeyP) {
		e.Paused = true
		e.Player.Speed = &geometry.Vector{}
	}

	if inp.IsKeyNewlyPressed(ebiten.KeyR) {
		e.Reset()
		return nil
	}

	if len(lo.Filter(e.Items, func(it *item.Item, _index int) bool {
		return !it.Collected && it.Important
	})) == 0 {
		e.IsWin = true
		return nil
	}

	if e.Player.IsDead() {
		return nil
	}

	e.ProcessPlayerInput(inp)
	e.Player.Move(&geometry.Vector{X: e.Player.Speed.X, Y: 0})
	e.AlignPlayerX()
	e.Player.Move(&geometry.Vector{X: 0, Y: e.Player.Speed.Y})
	e.AlignPlayerY()
	e.CheckPortals()
	e.CheckSpikes()
	e.CheckEnemyBullets()
	if err := e.CollectItems(); err != nil {
		return fmt.Errorf("collecting items: %w", err)
	}

	availableNPC := e.CheckNPCClose()
	if availableNPC != nil && inp.IsKeyNewlyPressed(ebiten.KeyE) {
		e.activeNPC = availableNPC
		e.activeNPC.Dialog.Greeting()
		return nil
	}

	e.Camera.MoveTo(e.Player.Origin.Add(&geometry.Vector{
		X: -camera.WIDTH/2 + e.Player.Width/2,
		Y: -camera.HEIGHT/2 + e.Player.Height/2,
	}))

	return nil
}

func (e *Engine) ProcessPlayerInput(inp *input.Input) {
	if e.Player.OnGround {
		e.Player.Acceleration.Y = 0
	} else {
		e.Player.Acceleration.Y = physics.GravityAcceleration
	}

	if (inp.IsKeyPressed(ebiten.KeySpace) || inp.IsKeyPressed(ebiten.KeyW)) && e.Player.OnGround {
		e.Player.Speed.Y = -5 * 2
	}

	switch {
	case inp.IsKeyPressed(ebiten.KeyA):
		e.Player.Speed.X = -2.5 * 2
		e.Player.LooksRight = false
	case inp.IsKeyPressed(ebiten.KeyD):
		e.Player.Speed.X = 2.5 * 2
		e.Player.LooksRight = true
	default:
		e.Player.Speed.X = 0
	}

	e.Player.ApplyAcceleration()
}

func (e *Engine) AlignPlayerX() {
	var pv *geometry.Vector

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.StaticTileType && c.Type() != object.InvWall {
			continue
		}

		pv = c.Rectangle().PushVectorX(e.Player.Rectangle())
		break
	}

	if pv == nil {
		return
	}

	e.Player.Move(pv)
}

func (e *Engine) AlignPlayerY() {
	var pv *geometry.Vector

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.StaticTileType && c.Type() != object.InvWall {
			continue
		}

		pv = c.Rectangle().PushVectorY(e.Player.Rectangle())
		break
	}

	e.Player.OnGround = false

	if pv == nil {
		return
	}

	e.Player.Move(pv)

	if pv.Y < 0 {
		e.Player.OnGround = true
	} else {
		e.Player.Speed.Y = 0
	}
}

func (e *Engine) CollectItems() error {
	collectedSomething := false

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Item {
			continue
		}

		it := c.(*item.Item)
		if it.Collected {
			continue
		}

		e.Player.Collect(it)

		collectedSomething = true
	}

	if collectedSomething {
		snapshot, err := e.MakeSnapshot()
		if err != nil {
			return fmt.Errorf("making snapshot: %w", err)
		}

		if err := e.SaveSnapshot(snapshot); err != nil {
			return fmt.Errorf("saving snapshot: %w", err)
		}
	}

	return nil
}

func (e *Engine) CheckPortals() {
	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Portal {
			continue
		}

		p := c.(*portal.Portal)
		if p.TeleportTo == nil {
			continue
		}

		e.Player.MoveTo(p.TeleportTo)
	}
}

func (e *Engine) CheckSpikes() {
	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Spike {
			continue
		}

		s := c.(*damage.Spike)
		e.Player.Health -= s.Damage
	}
}

func (e *Engine) CheckEnemyBullets() {
	var bullets []*damage.Bullet

	for _, b := range e.EnemyBullets {
		b.Move(b.Direction)
		ok := true
		for _, c := range e.Collisions(b.Rectangle()) {
			if c.Type() == object.StaticTileType {
				ok = false
				break
			}
		}
		if ok {
			bullets = append(bullets, b)
		}
	}

	e.EnemyBullets = bullets

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.EnemyBullet {
			continue
		}

		b := c.(*damage.Bullet)

		if b.Triggered {
			continue
		}

		e.Player.Health -= b.Damage
		b.Triggered = true
	}
}

func (e *Engine) CheckNPCClose() *npc.NPC {
	for _, c := range e.Collisions(e.Player.Rectangle().Extended(40)) {
		if c.Type() != object.NPC {
			continue
		}

		n := c.(*npc.NPC)
		return n
	}

	return nil
}

func (e *Engine) Checksum() (string, error) {
	b, err := msgpack.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("marshalling engine: %w", err)
	}
	if os.Getenv("DEBUG") == "1" {
		fmt.Println("==CHECKSUM==")
		fmt.Println(base64.StdEncoding.EncodeToString(b))
	}

	hash := sha256.New()
	if _, err := hash.Write(b); err != nil {
		return "", fmt.Errorf("hashing snapshot: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

var ErrInvalidChecksum = errors.New("invalid checksum")

func (e *Engine) ValidateChecksum(checksum string) error {
	if currentChecksum, err := e.Checksum(); err != nil {
		return fmt.Errorf("getting correct checksum: %w", err)
	} else if currentChecksum != checksum {
		return ErrInvalidChecksum
	}

	return nil
}

func (e *Engine) ActiveNPC() *npc.NPC {
	return e.activeNPC
}
