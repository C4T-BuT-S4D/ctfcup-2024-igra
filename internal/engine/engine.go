package engine

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/lafriks/go-tiled"
	"github.com/samber/lo"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/arcade"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/colors"
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
	Tiles            []*tiles.StaticTile      `json:"-" msgpack:"-"`
	Camera           *camera.Camera           `json:"-" msgpack:"camera"`
	Player           *player.Player           `json:"-" msgpack:"player"`
	Items            []*item.Item             `json:"items" msgpack:"items"`
	Portals          []*portal.Portal         `json:"-" msgpack:"portals"`
	Spikes           []*damage.Spike          `json:"-" msgpack:"spikes"`
	InvWalls         []*wall.InvWall          `json:"-" msgpack:"invWalls"`
	NPCs             []*npc.NPC               `json:"-" msgpack:"npcs"`
	Arcades          []*arcade.Machine        `json:"-" msgpack:"arcades"`
	EnemyBullets     []*damage.Bullet         `json:"-" msgpack:"enemyBullets"`
	BackgroundImages []*tiles.BackgroundImage `json:"-" msgpack:"backgroundImages"`

	StartSnapshot *Snapshot `json:"-" msgpack:"-"`

	fontsManager  *fonts.Manager
	spriteManager *sprites.Manager
	musicManager  *music.Manager
	snapshotsDir  string
	playerSpawn   *geometry.Point
	activeNPC     *npc.NPC
	activeArcade  *arcade.Machine
	dialogControl dialogControl

	Muted    bool   `json:"-" msgpack:"-"`
	Paused   bool   `json:"-" msgpack:"paused"`
	Tick     int    `json:"-" msgpack:"tick"`
	Level    string `json:"-" msgpack:"level"`
	IsWin    bool   `json:"-" msgpack:"isWin"`
	TeamName string `json:"-" msgpack:"-"`
}

var ErrNoPlayerSpawn = errors.New("no player spawn found")

func findPlayerSpawn(tileMap *tiled.Map) (*geometry.Point, error) {
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

type ResourceManager struct {
	Sprites *sprites.Manager
	Tiles   *tiles.Manager
	Fonts   *fonts.Manager
	Music   *music.Manager
}

func NewResourceManager(withMusic bool) *ResourceManager {
	rm := &ResourceManager{
		Sprites: sprites.NewManager(),
		Tiles:   tiles.NewManager(),
		Fonts:   fonts.NewManager(),
	}

	if withMusic {
		rm.Music = music.NewManager()
	}

	return rm
}

func New(config Config, resourceManager *ResourceManager, dialogProvider dialog.Provider, arcadeProvider arcade.Provider) (*Engine, error) {
	mapFile, err := resources.EmbeddedFS.Open(fmt.Sprintf("levels/%s.tmx", config.Level))
	if err != nil {
		return nil, fmt.Errorf("failed to open map: %w", err)
	}
	defer mapFile.Close()

	tmap, err := tiled.LoadReader("levels", mapFile, tiled.WithFileSystem(resources.EmbeddedFS))
	if err != nil {
		return nil, fmt.Errorf("failed to load map: %w", err)
	}

	var mapTiles []*tiles.StaticTile

	for _, l := range tmap.Layers {
		for x := 0; x < tmap.Width; x++ {
			for y := 0; y < tmap.Height; y++ {
				dt := l.Tiles[y*tmap.Width+x]
				if dt.IsNil() {
					continue
				}

				spriteRect := dt.Tileset.GetTileRect(dt.ID)

				if dt.Tileset.Image == nil {
					return nil, fmt.Errorf("tileset image is empty")
				}

				tilesImage := resourceManager.Tiles.Get(dt.Tileset.Image.Source)
				tileImage := tilesImage.SubImage(spriteRect).(*ebiten.Image)

				w, h := tmap.TileWidth, tmap.TileHeight
				mapTiles = append(
					mapTiles,
					tiles.NewStaticTile(
						&geometry.Point{
							X: float64(x * w),
							Y: float64(y * h),
						},
						w,
						h,
						tileImage,
					),
				)
			}
		}
	}

	var bgImages []*tiles.BackgroundImage
	for _, l := range tmap.ImageLayers {
		if l.Image == nil {
			return nil, fmt.Errorf("background image layer is empty")
		}

		bgImage := resourceManager.Tiles.Get(path.Base(l.Image.Source))
		bgImages = append(bgImages, &tiles.BackgroundImage{
			StaticTile: *tiles.NewStaticTile(
				&geometry.Point{
					X: float64(l.OffsetX),
					Y: float64(l.OffsetY),
				},
				l.Image.Width,
				l.Image.Height,
				bgImage),
		})
	}

	playerPos, err := findPlayerSpawn(tmap)
	if err != nil {
		return nil, fmt.Errorf("can't find player position: %w", err)
	}

	p, err := player.New(playerPos, resourceManager.Sprites)
	if err != nil {
		return nil, fmt.Errorf("creating player: %w", err)
	}

	var (
		items    []*item.Item
		spikes   []*damage.Spike
		invwalls []*wall.InvWall
		npcs     []*npc.NPC
		arcades  []*arcade.Machine
	)
	winPoints := make(map[string]*geometry.Point)
	portalsMap := make(map[string]*portal.Portal)

	for _, og := range tmap.ObjectGroups {
		for _, o := range og.Objects {
			switch o.Type {
			case "item":
				img := ebiten.NewImage(int(o.Width), int(o.Height))
				img.Fill(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff})

				if sprite := o.Properties.GetString("sprite"); sprite != "" {
					img = resourceManager.Sprites.GetSprite(sprites.Type(sprite))
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
					o.Properties.GetBool("important"),
				))
			case "portal":
				portalsMap[o.Name] = portal.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					resourceManager.Sprites.GetSprite(sprites.Portal),
					o.Width,
					o.Height,
					o.Properties.GetString("portal-to"),
					nil,
					o.Properties.GetString("boss"))
			case "spike":
				spikes = append(spikes, damage.NewSpike(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					resourceManager.Sprites.GetSprite(sprites.Spike),
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
				img := resourceManager.Sprites.GetSprite(sprites.Type(o.Properties.GetString("sprite")))
				dimg := resourceManager.Sprites.GetSprite(sprites.Type(o.Properties.GetString("dialog-sprite")))
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
					o.Properties.GetString("item"),
				))
			case "boss-win":
				winPoints[o.Name] = &geometry.Point{X: o.X, Y: o.Y}
			case "arcade":
				img := resourceManager.Sprites.GetSprite(sprites.Arcade)
				arc, err := arcadeProvider.Get(o.Name)
				if err != nil {
					return nil, fmt.Errorf("getting '%s' arcade: %w", o.Name, err)
				}
				arcades = append(arcades, arcade.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					o.Width,
					o.Height,
					arc,
					o.Properties.GetString("item"),
				))
			}
		}
	}

	for _, n := range npcs {
		i := slices.IndexFunc(items, func(i *item.Item) bool {
			return i.Name == n.ReturnsItem
		})
		if i < 0 {
			return nil, fmt.Errorf("item %s not found for npc", n.ReturnsItem)
		}
		n.LinkedItem = items[i]
	}
	for _, arc := range arcades {
		i := slices.IndexFunc(items, func(i *item.Item) bool {
			return i.Name == arc.ProvidesItem
		})
		if i < 0 {
			return nil, fmt.Errorf("item %s not found for arcade", arc.ProvidesItem)
		}
		arc.LinkedItem = items[i]
	}

	for name, p := range portalsMap {
		if p.PortalTo == "" {
			continue
		}
		toPortal := portalsMap[p.PortalTo]
		if toPortal == nil {
			return nil, fmt.Errorf("destination %s not found for portal %s", p.PortalTo, name)
		}
		//p.TeleportTo = toPortal.Origin.Add(&geometry.Vector{
		//	X: 32,
		//	Y: 0,
		//})
		p.TeleportTo = toPortal.Origin
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
		Tiles:            mapTiles,
		BackgroundImages: bgImages,
		Camera:           cam,
		Player:           p,
		Items:            items,
		Portals:          portals,
		Spikes:           spikes,
		InvWalls:         invwalls,
		NPCs:             npcs,
		Arcades:          arcades,
		spriteManager:    resourceManager.Sprites,
		fontsManager:     resourceManager.Fonts,
		musicManager:     resourceManager.Music,
		snapshotsDir:     config.SnapshotsDir,
		playerSpawn:      playerPos,
		Level:            config.Level,
		TeamName:         strings.Split(os.Getenv("AUTH_TOKEN"), ":")[0],
		dialogControl: dialogControl{
			maskInput: !dialogProvider.DisplayInput(),
		},
	}, nil
}

func NewFromSnapshot(config Config, snapshot *Snapshot, resourceManager *ResourceManager, dialogProvider dialog.Provider, arcadeProvider arcade.Provider) (*Engine, error) {
	e, err := New(config, resourceManager, dialogProvider, arcadeProvider)
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
	e.activeArcade = nil
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

	width, _ := text.Measure("YOU DIED", face, 0)

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/2)
	textOp.ColorScale.ScaleWithColor(redColor)
	text.Draw(screen, "YOU DIED", face, textOp)
}

func (e *Engine) drawYouWinScreen(screen *ebiten.Image) {
	face := e.fontsManager.Get(fonts.DSouls)
	gColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	width, _ := text.Measure("YOU WIN", face, 0)

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/2)
	textOp.ColorScale.ScaleWithColor(gColor)
	text.Draw(screen, "YOU WIN", face, textOp)
}

func (e *Engine) drawArcadeState(screen *ebiten.Image) {
	as := e.activeArcade.Game.State()

	borderSizePx := 10
	cameraRectSide := min(camera.WIDTH, camera.HEIGHT)
	cameraInnerRectSide := cameraRectSide - borderSizePx*2

	scaleFactor := cameraInnerRectSide / arcade.ScreenSize

	mgx, mgy := float32((camera.WIDTH-cameraRectSide)/2), float32((camera.HEIGHT-cameraRectSide)/2)
	borderSizeF := float32(borderSizePx)
	vector.DrawFilledRect(screen, mgx, mgy, float32(cameraRectSide), float32(cameraRectSide), color.White, false)
	vector.DrawFilledRect(screen, mgx+borderSizeF, mgy+borderSizeF, float32(cameraInnerRectSide), float32(cameraInnerRectSide), color.Black, false)

	for i := 0; i < len(as.Screen); i++ {
		for j := 0; j < len(as.Screen[i]); j++ {
			dy := float32(i * scaleFactor)
			dx := float32(j * scaleFactor)
			vector.DrawFilledRect(screen, mgx+dx+borderSizeF, mgy+dy+borderSizeF, float32(scaleFactor), float32(scaleFactor), as.Screen[i][j], false)
		}
	}

	if as.Result == arcade.ResultUnknown {
		return
	}

	txt, txtC := "TRY AGAIN", colors.Red
	if as.Result == arcade.ResultWon {
		txt, txtC = "YOU WIN. PRESS ESC TO CONTINUE", colors.Green
	}

	face := e.fontsManager.Get(fonts.DSouls)
	width, _ := text.Measure(txt, face, 0)
	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/2)
	textOp.ColorScale.ScaleWithColor(txtC)
	text.Draw(screen, txt, face, textOp)
}

func (e *Engine) drawNPCDialog(screen *ebiten.Image) {

	// Draw dialog border (outer rectangle).
	borderw, borderh := camera.WIDTH-camera.WIDTH/8, camera.HEIGHT/2
	bx, by := float32(camera.WIDTH/16.0), float32(camera.HEIGHT/2.0-camera.HEIGHT/16)
	vector.DrawFilledRect(screen, bx, by, float32(borderw), float32(borderh), color.White, false)

	// Draw dialog border (inner rectangle).
	ibw, ibh := borderw-camera.WIDTH/32, borderh-camera.HEIGHT/32
	ibx, iby := bx+camera.WIDTH/64, by+camera.HEIGHT/64
	vector.DrawFilledRect(screen, ibx, iby, float32(ibw), float32(ibh), color.Black, false)

	// Draw dialog NPC image.
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(camera.WIDTH/2+camera.WIDTH/8, camera.HEIGHT/2)
	screen.DrawImage(e.activeNPC.DialogImage, op)

	// Draw dialog text.
	dtx, dty := float64(ibx+camera.WIDTH/32), float64(iby+camera.HEIGHT/32)
	face := e.fontsManager.Get(fonts.Dialog)
	faceMetrics := face.Metrics()
	lineHeight := faceMetrics.HAscent + faceMetrics.HDescent
	txt := e.activeNPC.Dialog.State().Text

	lines := input.AutoWrap(txt, face, ibw-camera.WIDTH/32)
	e.dialogControl.scroll = max(min(e.dialogControl.scroll, len(lines)-1), 0)

	l := e.dialogControl.scroll
	r := min(e.dialogControl.scroll+dialogShowLines, len(lines))

	visibleLines := lines[l:r]
	textOp := &text.DrawOptions{LayoutOptions: text.LayoutOptions{LineSpacing: lineHeight}}
	textOp.GeoM.Translate(dtx, dty)
	textOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, strings.Join(visibleLines, "\n"), face, textOp)

	// Draw dialog input buffer.
	if len(e.dialogControl.inputBuffer) > 0 {
		dtbx, dtby := dtx, dty+float64(len(visibleLines))*lineHeight
		ibuf := string(e.dialogControl.inputBuffer)
		if e.dialogControl.maskInput {
			ibuf = strings.Repeat("*", len(ibuf))
		}
		x := input.AutoWrap(ibuf, face, ibw-camera.WIDTH/32)

		textOp := &text.DrawOptions{LayoutOptions: text.LayoutOptions{LineSpacing: lineHeight}}
		textOp.GeoM.Translate(dtbx, dtby)
		textOp.ColorScale.ScaleWithColor(color.RGBA{R: 0x00, G: 0xff, B: 0xff, A: 0xff})
		text.Draw(screen, strings.Join(x, "\n"), face, textOp)
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
		case object.BackgroundImage:
			bi := c.(*tiles.BackgroundImage)
			screen.DrawImage(bi.Image, op)
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
		case object.Arcade:
			a := c.(*arcade.Machine)
			screen.DrawImage(a.Image, op)
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
		start := float64(16)
		step := float64(36)
		textOp := &text.DrawOptions{}
		textOp.GeoM.Translate(start, start)
		textOp.ColorScale.ScaleWithColor(color.RGBA{R: 204, G: 14, B: 206, A: 255})
		text.Draw(screen, teamtxt, face, textOp)

		txt := fmt.Sprintf("HP: %d", e.Player.Health)
		textOp = &text.DrawOptions{}
		textOp.GeoM.Translate(start, start+step*1)
		textOp.ColorScale.ScaleWithColor(color.RGBA{R: 0, G: 255, B: 0, A: 255})
		text.Draw(screen, txt, face, textOp)

		tickTxt := fmt.Sprintf("Tick: %d", e.Tick)
		textOp = &text.DrawOptions{}
		textOp.GeoM.Translate(start, start+step*2)
		textOp.ColorScale.ScaleWithColor(color.RGBA{R: 0, G: 255, B: 0, A: 255})
		text.Draw(screen, tickTxt, face, textOp)

		for i, it := range e.Player.Inventory.Items {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(e.Camera.Width-float64(i+1)*72, 72)
			screen.DrawImage(it.Image, op)
		}
	}

	if e.activeNPC != nil {
		e.drawNPCDialog(screen)
	}

	if e.activeArcade != nil {
		e.drawArcadeState(screen)
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

	if e.activeArcade != nil {
		if inp.IsKeyNewlyPressed(ebiten.KeyEscape) {
			if err := e.activeArcade.Game.Stop(); err != nil {
				return fmt.Errorf("stopping arcade game: %w", err)
			}
			e.activeArcade = nil
			return nil
		}
		if e.activeArcade.Game.State().Won && e.activeArcade.LinkedItem != nil {
			e.activeArcade.LinkedItem.MoveTo(e.activeArcade.Origin.Add(&geometry.Vector{
				X: +64,
				Y: +32,
			}))
			return nil
		}
		if e.activeArcade.Game.State().Result != arcade.ResultUnknown {
			// No need to feed the game if the result is known.
			return nil
		}
		if e.Tick%5 != 0 {
			return nil
		}
		if err := e.activeArcade.Game.Feed(inp.PressedKeys()); err != nil {
			return fmt.Errorf("feeding arcade game: %w", err)
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

	availableArcade := e.CheckArcadeClose()
	if availableArcade != nil && inp.IsKeyNewlyPressed(ebiten.KeyE) {
		e.activeArcade = availableArcade
		if err := e.activeArcade.Game.Start(); err != nil {
			return fmt.Errorf("starting arcade game: %w", err)
		}
	}

	e.Camera.MoveTo(e.Player.Origin.Add(&geometry.Vector{
		X: -camera.WIDTH/2 + e.Player.Width/2,
		Y: -camera.HEIGHT/2 + e.Player.Height/2,
	}))

	return nil
}

func (e *Engine) ProcessPlayerInput(inp *input.Input) {
	if e.Player.OnGround() {
		e.Player.Acceleration.Y = 0
	} else {
		e.Player.Acceleration.Y = physics.GravityAcceleration
	}

	if (inp.IsKeyPressed(ebiten.KeySpace) || inp.IsKeyPressed(ebiten.KeyW)) && e.Player.OnGroundCoyote() {
		e.Player.Speed.Y = -5 * 2
		e.Player.ResetCoyote()
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

	for _, c := range e.Collisions(e.Player.Rectangle(), object.StaticTileType, object.InvWall) {
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

	for _, c := range e.Collisions(e.Player.Rectangle(), object.StaticTileType, object.InvWall) {
		pv = c.Rectangle().PushVectorY(e.Player.Rectangle())
		break
	}

	e.Player.SetOnGround(false, e.Tick)

	if pv == nil {
		return
	}

	e.Player.Move(pv)

	if pv.Y < 0 {
		e.Player.SetOnGround(true, e.Tick)
	} else {
		e.Player.Speed.Y = 0
	}
}

func (e *Engine) CollectItems() error {
	collectedSomething := false

	for _, c := range e.Collisions(e.Player.Rectangle(), object.Item) {
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
	for _, c := range e.Collisions(e.Player.Rectangle(), object.Portal) {
		p := c.(*portal.Portal)
		if p.TeleportTo == nil {
			continue
		}

		dx := 32.0
		if e.Player.Speed.X < 0 {
			e.Player.MoveTo(p.TeleportTo.Add(&geometry.Vector{
				X: -dx,
				Y: 0,
			}))
		} else {
			e.Player.MoveTo(p.TeleportTo.Add(&geometry.Vector{
				X: dx,
				Y: 0,
			}))
		}
	}
}

func (e *Engine) CheckSpikes() {
	for _, c := range e.Collisions(e.Player.Rectangle(), object.Spike) {
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

	for _, c := range e.Collisions(e.Player.Rectangle(), object.EnemyBullet) {
		b := c.(*damage.Bullet)

		if b.Triggered {
			continue
		}

		e.Player.Health -= b.Damage
		b.Triggered = true
	}
}

func (e *Engine) CheckNPCClose() *npc.NPC {
	for _, c := range e.Collisions(e.Player.Rectangle().Extended(40), object.NPC) {
		n := c.(*npc.NPC)
		return n
	}

	return nil
}

func (e *Engine) CheckArcadeClose() *arcade.Machine {
	for _, c := range e.Collisions(e.Player.Rectangle().Extended(40), object.Arcade) {
		a := c.(*arcade.Machine)
		return a
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
