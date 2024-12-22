package engine

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"

	// Register png codec.
	_ "image/png"
	"math/rand/v2"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/boss"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/casino"
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
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/input"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/npc"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/physics"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/platform"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/player"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/portal"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/tiles"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"
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
	Camera           *camera.Camera           `json:"-" msgpack:"-"`
	Player           *player.Player           `json:"-" msgpack:"player"`
	Items            []*item.Item             `json:"items" msgpack:"items"`
	Portals          []*portal.Portal         `json:"-" msgpack:"portals"`
	Spikes           []*damage.Spike          `json:"-" msgpack:"spikes"`
	Platforms        []*platform.Platform     `json:"-" msgpack:"platforms"`
	NPCs             []*npc.NPC               `json:"-" msgpack:"npcs"`
	Arcades          []*arcade.Machine        `json:"-" msgpack:"arcades"`
	Slots            []*casino.SlotMachine    `json:"-" msgpack:"slots"`
	EnemyBullets     []*damage.Bullet         `json:"-" msgpack:"enemyBullets"`
	BackgroundImages []*tiles.BackgroundImage `json:"-" msgpack:"backgroundImages"`

	BossEntered  bool            `json:"-" msgpack:"bossEntered"`
	Boss         boss.BOSS       `json:"-" msgpack:"boss"`
	BossItem     *item.Item      `json:"-" msgpack:"bossItem"`
	BossPortal   *portal.Portal  `json:"-" msgpack:"bossPortal"`
	BossWinPoint *geometry.Point `json:"-" msgpack:"bossWinPoint"`

	StartSnapshot *Snapshot `json:"-" msgpack:"-"`

	resourceBundle      *resources.Bundle
	snapshotsDir        string
	playerSpawn         geometry.Point
	activeNPC           *npc.NPC
	activeArcade        *arcade.Machine
	dialogControl       dialogControl
	notificationText    string
	notificationEndTick int

	Muted    bool   `json:"-" msgpack:"-"`
	Paused   bool   `json:"-" msgpack:"paused"`
	Tick     int    `json:"-" msgpack:"tick"`
	Level    string `json:"-" msgpack:"level"`
	IsWin    bool   `json:"-" msgpack:"isWin"`
	TeamName string `json:"-" msgpack:"-"`
}

var ErrNoPlayerSpawn = errors.New("no player spawn found")

func findPlayerSpawn(tileMap *tiled.Map) (geometry.Point, error) {
	for _, og := range tileMap.ObjectGroups {
		for _, o := range og.Objects {
			if o.Type == "player_spawn" {
				return geometry.Point{
					X: o.X,
					Y: o.Y,
				}, nil
			}
		}
	}

	return geometry.Point{}, ErrNoPlayerSpawn
}

func New(config Config, resourceBundle *resources.Bundle, dialogProvider dialog.Provider, arcadeProvider arcade.Provider) (*Engine, error) {
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
	var backgroundTiles []*tiles.BackgroundImage

	for _, l := range tmap.Layers {
		collisions := l.Properties.GetBool("collisions")
		for y := range tmap.Height {
			for x := range tmap.Width {
				dt := l.Tiles[y*tmap.Width+x]
				if dt.IsNil() {
					continue
				}

				if dt.Tileset.Image == nil {
					return nil, fmt.Errorf("tileset image is empty")
				}

				spriteRect := dt.Tileset.GetTileRect(dt.ID)
				tilesImage := resourceBundle.GetTile(dt.Tileset.Image.Source)
				tileImage := tilesImage.SubImage(spriteRect).(*ebiten.Image)

				tile := tiles.NewStaticTile(
					geometry.Point{
						X: float64(x * tmap.TileWidth),
						Y: float64(y * tmap.TileHeight),
					},
					tmap.TileWidth,
					tmap.TileHeight,
					tileImage,
					tiles.Flips{
						Horizontal: dt.HorizontalFlip,
						Vertical:   dt.VerticalFlip,
						Diagonal:   dt.DiagonalFlip,
					},
				)

				if collisions {
					mapTiles = append(mapTiles, tile)
				} else {
					backgroundTiles = append(backgroundTiles, &tiles.BackgroundImage{StaticTile: *tile})
				}
			}
		}
	}

	var bgImages []*tiles.BackgroundImage
	for _, l := range tmap.ImageLayers {
		if l.Image == nil {
			return nil, fmt.Errorf("background image layer is empty")
		}

		bgImage := resourceBundle.GetTile(path.Base(l.Image.Source))
		bgImages = append(bgImages, &tiles.BackgroundImage{
			StaticTile: *tiles.NewStaticTile(
				geometry.Point{
					X: float64(l.OffsetX),
					Y: float64(l.OffsetY),
				},
				l.Image.Width,
				l.Image.Height,
				bgImage,
				tiles.Flips{},
			),
		})
	}

	playerPos, err := findPlayerSpawn(tmap)
	if err != nil {
		return nil, fmt.Errorf("can't find player position: %w", err)
	}

	p, err := player.New(playerPos, resourceBundle.SpriteBundle)
	if err != nil {
		return nil, fmt.Errorf("creating player: %w", err)
	}

	var (
		items         []*item.Item
		spikes        []*damage.Spike
		platforms     []*platform.Platform
		npcs          []*npc.NPC
		arcades       []*arcade.Machine
		slots         []*casino.SlotMachine
		bossObj       boss.BOSS
		bossItemName  string
		bossItem      *item.Item
		bossPortal    *portal.Portal
		bossWinPoint  *geometry.Point
		bossPlatforms []*platform.Platform
	)
	portalsMap := make(map[string]*portal.Portal)

	for _, og := range tmap.ObjectGroups {
		for _, o := range og.Objects {
			switch o.Type {
			case "item":
				img := ebiten.NewImage(int(o.Width), int(o.Height))
				img.Fill(color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xff})

				if sprite := o.Properties.GetString("sprite"); sprite != "" {
					img = resourceBundle.GetSprite(resources.SpriteType(sprite))
				}

				items = append(items, item.New(
					geometry.Point{
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
				s := resources.SpritePortal
				if sprite := o.Properties.GetString("sprite"); sprite != "" {
					s = resources.SpriteType(sprite)
				}

				portalsMap[o.Name] = portal.New(
					geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					resourceBundle.GetSprite(s),
					o.Width,
					o.Height,
					o.Properties.GetString("portal-to"),
					o.Properties.GetString("boss"),
				)
			case "spike":
				spikes = append(spikes, damage.NewSpike(
					geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					resourceBundle.GetDirectionalSprite("spike", o.Properties.GetString("direction")),
					o.Width,
					o.Height,
				))
			case "moving_spike":
				spikes = append(spikes, damage.NewMovingSpike(
					geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					o.Width,
					o.Height,
					resourceBundle.GetDirectionalSprite("spike", o.Properties.GetString("direction")),
					physics.ParsePath(o.Properties.GetString("path")),
					o.Properties.GetInt("distance"),
					o.Properties.GetInt("speed"),
				))
			case "platform":
				sprite := lo.
					If(strings.HasPrefix(o.Name, "boss"), resources.SpritePlatformWide).
					Else(resources.SpritePlatform)
				platforms = append(platforms, platform.New(
					geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					o.Width,
					o.Height,
					resourceBundle.GetSprite(sprite),
					physics.ParsePath(o.Properties.GetString("path")),
					o.Properties.GetInt("distance"),
					o.Properties.GetInt("speed"),
				))
				if strings.HasPrefix(o.Name, "boss") {
					bossPlatforms = append(bossPlatforms, platforms[len(platforms)-1])
				}
			case "npc":
				img := resourceBundle.GetSprite(resources.SpriteType(o.Properties.GetString("sprite")))
				dimg := resourceBundle.GetSprite(resources.SpriteType(o.Properties.GetString("dialog-sprite")))
				npcd, err := dialogProvider.Get(o.Name)
				if err != nil {
					return nil, fmt.Errorf("getting '%s' dialog: %w", o.Name, err)
				}
				npcs = append(npcs, npc.New(
					geometry.Point{
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
				bossWinPoint = &geometry.Point{X: o.X, Y: o.Y}
			case "boss":
				img := resourceBundle.GetSprite(resources.SpriteType(o.Properties.GetString("sprite")))
				bulletImg := resourceBundle.GetSprite(resources.SpriteBullet)
				if o.Name == "v2" {
					bossObj = boss.NewV2(
						object.NewRendered(
							geometry.Point{
								X: o.X,
								Y: o.Y,
							},
							img,
							o.Width,
							o.Height,
						),
						bulletImg,
					)
				} else {
					bossObj = boss.NewV1(
						object.NewRendered(
							geometry.Point{
								X: o.X,
								Y: o.Y,
							},
							img,
							o.Width,
							o.Height,
						),
						bulletImg,
					)
				}

				bossItemName = o.Properties.GetString("item")
			case "slots":
				img := resourceBundle.GetSprite(resources.SpriteType(o.Properties.GetString("sprite")))
				slotMachine := casino.NewSlotMachine(
					geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					o.Width,
					o.Height,
					o.Properties.GetInt("coins"),
					o.Properties.GetInt("cooldown"),
					o.Properties.GetInt("seed"),
					o.Properties.GetFloat("probability"),
				)
				slots = append(slots, slotMachine)
			case "arcade":
				img := resourceBundle.GetSprite(resources.SpriteArcade)
				arc, err := arcadeProvider.Get(o.Name)
				if err != nil {
					return nil, fmt.Errorf("getting '%s' arcade: %w", o.Name, err)
				}
				arcades = append(arcades, arcade.New(
					geometry.Point{
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

	if bossItemName != "" {
		bossItem, _ = lo.Find(items, func(i *item.Item) bool {
			return i.Name == bossItemName
		})
		bossPortal = portalsMap["boss-exit"]
	}

	if b, ok := bossObj.(*boss.V2); ok {
		b.SetPlatforms(bossPlatforms)
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

		toPortal, ok := portalsMap[p.PortalTo]
		if !ok {
			return nil, fmt.Errorf("destination %s not found for portal %s", p.PortalTo, name)
		}

		p.TeleportTo = toPortal.Origin
	}

	cam := &camera.Camera{
		Base: &object.Base{
			Origin: geometry.Point{
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
		BackgroundImages: slices.Concat(bgImages, backgroundTiles),
		Camera:           cam,
		Player:           p,
		Items:            items,
		Portals:          portals,
		Spikes:           spikes,
		Platforms:        platforms,
		NPCs:             npcs,
		Boss:             bossObj,
		BossItem:         bossItem,
		BossPortal:       bossPortal,
		BossWinPoint:     bossWinPoint,
		Arcades:          arcades,
		Slots:            slots,
		resourceBundle:   resourceBundle,
		snapshotsDir:     config.SnapshotsDir,
		playerSpawn:      playerPos,
		Level:            config.Level,
		TeamName:         strings.Split(os.Getenv("AUTH_TOKEN"), ":")[0],
		dialogControl: dialogControl{
			maskInput: !dialogProvider.DisplayInput(),
		},
	}, nil
}

func NewFromSnapshot(config Config, snapshot *Snapshot, resourceBundle *resources.Bundle, dialogProvider dialog.Provider, arcadeProvider arcade.Provider) (*Engine, error) {
	e, err := New(config, resourceBundle, dialogProvider, arcadeProvider)
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}

	e.StartSnapshot = snapshot
	itemMap := make(map[string]*item.Item)
	for _, it := range snapshot.Items {
		itemMap[it.Name] = it
	}
	// TODO: use copier later if needed.
	for _, it := range e.Items {
		if sit, ok := itemMap[it.Name]; ok {
			it.Collected = sit.Collected
			it.Important = sit.Important
		}
	}

	for _, it := range e.Items {
		if it.Collected {
			e.Player.Inventory.Items = append(e.Player.Inventory.Items, it)
		}
	}

	return e, nil
}

type Snapshot struct {
	Items     []*item.Item `json:"items"`
	CreatedAt time.Time    `json:"created_at"`
}

func NewSnapshotFromProto(proto *gameserverpb.EngineSnapshot) (*Snapshot, error) {
	var s Snapshot
	if err := s.FromJSON(proto.Data); err != nil {
		return nil, fmt.Errorf("unmarshalling snapshot: %w", err)
	}
	return &s, nil
}

func (s *Snapshot) FromJSON(data []byte) error {
	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("unmarshalling snapshot: %w", err)
	}
	return nil
}

func (s *Snapshot) ToJSON() ([]byte, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshalling snapshot: %w", err)
	}
	return data, nil
}

func (s *Snapshot) ToProto() (*gameserverpb.EngineSnapshot, error) {
	if s == nil {
		return nil, nil
	}
	serialized, err := s.ToJSON()
	if err != nil {
		return nil, err
	}
	return &gameserverpb.EngineSnapshot{
		Data: serialized,
	}, nil
}

func (e *Engine) Reset() {
	e.Player.MoveTo(e.playerSpawn)
	e.Player.Health = player.DefaultHealth
	*e.Player.Physical = physics.Physical{}
	e.activeNPC = nil
	e.activeArcade = nil
	for _, sm := range e.Slots {
		sm.Reset()
	}
	e.notificationText = ""
	e.notificationEndTick = 0
	e.EnemyBullets = nil
	e.Tick = 0
	e.Player.ResetCoyote()

	e.BossEntered = false
	if e.Boss != nil {
		e.Boss.Reset()
	}
}

func (e *Engine) MakeSnapshot() *Snapshot {
	return &Snapshot{
		Items:     e.Items,
		CreatedAt: time.Now().UTC(),
	}
}

func (e *Engine) SaveSnapshot(snapshot *Snapshot) error {
	if e.snapshotsDir == "" {
		return nil
	}

	data, err := snapshot.ToJSON()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("snapshot_%s_%s", e.Level, snapshot.CreatedAt.Format("2006-01-02T15:04:05.999999999"))

	if err := os.WriteFile(filepath.Join(e.snapshotsDir, filename), data, 0o400); err != nil {
		return fmt.Errorf("writing snapshot file: %w", err)
	}

	return nil
}

func (e *Engine) soulsFont() text.Face {
	return e.resourceBundle.GetFontFace(resources.FontSouls, camera.WIDTH/16)
}

func (e *Engine) dialogFont() text.Face {
	return e.resourceBundle.GetFontFace(resources.FontDialog, camera.WIDTH/40)
}

func (e *Engine) drawDiedScreen(screen *ebiten.Image) {
	redColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	face := e.soulsFont()

	width, _ := text.Measure("YOU DIED", face, 0)

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/2)
	textOp.ColorScale.ScaleWithColor(redColor)
	text.Draw(screen, "YOU DIED", face, textOp)
}

func (e *Engine) drawYouWinScreen(screen *ebiten.Image) {
	face := e.soulsFont()
	gColor := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	width, _ := text.Measure("YOU WIN", face, 0)

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/2)
	textOp.ColorScale.ScaleWithColor(gColor)
	text.Draw(screen, "YOU WIN", face, textOp)
}

func (e *Engine) drawNotification(screen *ebiten.Image) {
	if e.notificationText == "" {
		return
	}

	face := e.dialogFont()
	yellow := color.RGBA{R: 255, G: 255, B: 0, A: 255}

	width, _ := text.Measure(e.notificationText, face, 0)

	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/8)
	textOp.ColorScale.ScaleWithColor(yellow)
	text.Draw(screen, e.notificationText, face, textOp)
}

func (e *Engine) drawArcadeState(screen *ebiten.Image) {
	as := e.activeArcade.Game.State()

	borderSizePx := 10
	cameraInnerRectSide := (min(camera.WIDTH, camera.HEIGHT) - borderSizePx*2) / arcade.ScreenSize * arcade.ScreenSize
	cameraRectSide := cameraInnerRectSide + borderSizePx*2

	scaleFactor := cameraInnerRectSide / arcade.ScreenSize

	mgx, mgy := float32((camera.WIDTH-cameraRectSide)/2), float32((camera.HEIGHT-cameraRectSide)/2)
	borderSizeF := float32(borderSizePx)
	vector.DrawFilledRect(screen, mgx, mgy, float32(cameraRectSide), float32(cameraRectSide), color.White, false)
	vector.DrawFilledRect(screen, mgx+borderSizeF, mgy+borderSizeF, float32(cameraInnerRectSide), float32(cameraInnerRectSide), color.Black, false)

	for i := range arcade.ScreenSize {
		for j := range arcade.ScreenSize {
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

	face := e.soulsFont()
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
	iw, ih := e.activeNPC.DialogImage.Bounds().Dx(), e.activeNPC.DialogImage.Bounds().Dy()
	wantHeight := camera.HEIGHT / 2
	scaleH := float64(wantHeight) / float64(ih)
	wantWidth := int(float64(iw) * scaleH)
	tx, ty := camera.WIDTH/2+camera.WIDTH/8, camera.HEIGHT/2-camera.HEIGHT/16

	op := &ebiten.DrawImageOptions{}
	scaledImg := ebiten.NewImage(wantWidth, wantHeight)
	op.GeoM.Scale(scaleH, scaleH)
	scaledImg.DrawImage(e.activeNPC.DialogImage, op)
	op.GeoM.Reset()
	op.GeoM.Translate(float64(tx), float64(ty))
	screen.DrawImage(scaledImg, op)

	// Draw dialog text.
	dtx, dty := float64(ibx+camera.WIDTH/32), float64(iby+camera.HEIGHT/32)
	face := e.dialogFont()
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

	if e.activeArcade == nil {
		for _, c := range e.Collisions(e.Camera.Rectangle()) {
			visible := c.Rectangle().Sub(e.Camera.Rectangle())
			base := geometry.Origin.Add(visible)
			op := &ebiten.DrawImageOptions{}

			switch o := c.(type) {
			case *player.Player:
				if e.Player.LooksRight {
					op.GeoM.Scale(-1, 1)
					op.GeoM.Translate(e.Player.Width, 0)
				}
			case *tiles.StaticTile:
				o.ApplyFlips(op)
			case *tiles.BackgroundImage:
				o.ApplyFlips(op)
			case *damage.Bullet:
				op.GeoM.Scale(4, 4)
				op.GeoM.Translate(-2, 0)
			default:
				// not a player or boss.
			}

			op.GeoM.Translate(
				base.X,
				base.Y,
			)

			switch obj := c.(type) {
			case *item.Item:
				if !obj.Collected {
					screen.DrawImage(obj.Image(), op)
				}
			case *damage.Bullet:
				if !obj.Triggered {
					screen.DrawImage(obj.Image(), op)
				}
			case object.Drawable:
				screen.DrawImage(obj.Image(), op)
			default:
			}
		}
	}

	if e.BossEntered && e.Boss != nil {
		bossHealth := e.Boss.Health()
		if bossHealth != nil && bossHealth.Health > 0 {
			op := &ebiten.DrawImageOptions{}
			width := float64(camera.WIDTH) * float64(bossHealth.Health) / float64(bossHealth.MaxHealth)
			op.GeoM.Scale(width, camera.HEIGHT/30)
			op.GeoM.Translate((float64(camera.WIDTH)-width)/2, 0)

			bossHpImage := e.resourceBundle.GetSprite(resources.SpriteHP)
			screen.DrawImage(bossHpImage, op)
		}
	}

	if !e.Player.IsDead() {
		face := e.dialogFont()
		start := float64(camera.WIDTH / 80)
		step := float64(camera.WIDTH / 32)
		index := float64(0)

		teamtxt := fmt.Sprintf("Team %s", e.TeamName)
		textOp := &text.DrawOptions{}
		textOp.GeoM.Translate(start, start+step*index)
		textOp.ColorScale.ScaleWithColor(color.RGBA{R: 204, G: 14, B: 206, A: 255})
		text.Draw(screen, teamtxt, face, textOp)
		index++

		if e.activeArcade == nil {
			txt := fmt.Sprintf("HP: %d", e.Player.Health)
			textOp = &text.DrawOptions{}
			textOp.GeoM.Translate(start, start+step*index)
			textOp.ColorScale.ScaleWithColor(color.RGBA{R: 0, G: 255, B: 0, A: 255})
			text.Draw(screen, txt, face, textOp)
			index++
		}

		coinsTxt := fmt.Sprintf("Coins: %d", e.Player.Coins)
		textOp = &text.DrawOptions{}
		textOp.GeoM.Translate(start, start+step*index)
		textOp.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 215, B: 0, A: 255})
		text.Draw(screen, coinsTxt, face, textOp)

		for i, it := range e.Player.Inventory.Items {
			itemX := e.Camera.Width - float64(i+1)*72
			itemY := camera.HEIGHT / 20
			vector.StrokeRect(screen, float32(itemX-4), float32(itemY-4), 40, 40, 2, color.RGBA{R: 230, G: 230, B: 230, A: 255}, false)
			vector.DrawFilledRect(screen, float32(itemX-4), float32(itemY-4), 40, 40, color.RGBA{R: 156, G: 150, B: 138, A: 196}, false)

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(itemX), float64(itemY))
			screen.DrawImage(it.Image(), op)
		}
	}

	if e.activeNPC != nil {
		e.drawNPCDialog(screen)
	}

	if e.activeArcade != nil {
		e.drawArcadeState(screen)
	}

	e.drawNotification(screen)
}

func (e *Engine) Update(inp *input.Input) error {
	e.Tick++

	if e.resourceBundle.MusicBundle != nil {
		p := e.resourceBundle.GetMusicPlayer(resources.MusicBackground)
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
		if e.activeNPC.Dialog.State().GaveItem && e.activeNPC.LinkedItem != nil {
			e.activeNPC.LinkedItem.MoveTo(e.Player.Origin)
			e.activeNPC.LinkedItem = nil
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
				e.activeNPC.Dialog.Feed(string(e.dialogControl.inputBuffer), e.Player.Coins)
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

		if inp.IsKeyNewlyPressed(ebiten.KeyR) {
			if err := e.activeArcade.Game.Stop(); err != nil {
				return fmt.Errorf("stopping arcade game: %w", err)
			}
			if err := e.activeArcade.Game.Start(); err != nil {
				return fmt.Errorf("restarting arcade game: %w", err)
			}
			return nil
		}

		if result := e.activeArcade.Game.State().Result; result == arcade.ResultWon && e.activeArcade.LinkedItem != nil {
			e.activeArcade.LinkedItem.MoveTo(e.Player.Origin)
			e.activeArcade.LinkedItem = nil
			return nil
		} else if result != arcade.ResultUnknown {
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

	e.ProcessMovingX()
	e.Player.ApplyAccelerationX()
	e.Player.Move(geometry.Vector{X: e.Player.Speed.X, Y: 0})
	e.AlignPlayerX()

	e.ProcessMovingY()
	e.Player.ApplyAccelerationY()
	e.Player.Move(geometry.Vector{X: 0, Y: e.Player.Speed.Y})
	e.AlignPlayerY()

	e.CheckPortals()
	e.CheckSpikes()
	e.CheckEnemyBullets()
	e.CheckBoss()
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

	availableSlots := e.CheckSlotsClose()
	if availableSlots != nil && inp.IsKeyNewlyPressed(ebiten.KeyE) {
		e.PlaySlots(availableSlots)
	}

	if e.notificationEndTick > 0 && e.Tick >= e.notificationEndTick {
		e.notificationText = ""
		e.notificationEndTick = 0
	}

	e.Camera.MoveTo(e.Player.Origin.Add(geometry.Vector{
		X: -camera.WIDTH/2 + e.Player.Width/2,
		Y: -camera.HEIGHT/2 + e.Player.Height/2,
	}))

	return nil
}

func (e *Engine) ProcessPlayerInput(inp *input.Input) {
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
}

func (e *Engine) ProcessMovingX() {
	for _, s := range e.Spikes {
		s.MoveX()
	}

	for _, p := range e.Platforms {
		p.MoveX()
	}

	if p, ok := e.Player.OnGround().(*platform.Platform); ok {
		e.Player.Origin.X += p.Speed.X
	}
}

func (e *Engine) ProcessMovingY() {
	for _, s := range e.Spikes {
		s.MoveY()
	}

	for _, p := range e.Platforms {
		p.MoveY()
	}

	if p, ok := e.Player.OnGround().(*platform.Platform); ok {
		e.Player.Acceleration.Y += p.Acceleration.Y
	}
}

func (e *Engine) AlignPlayerX() {
	var pv geometry.Vector
	var pvOk bool

	for t := range Collide2(e.Player.Rectangle(), e.Platforms, e.Tiles) {
		pv, pvOk = t.Rectangle().PushVectorX(e.Player.Rectangle()), true
		break
	}

	if !pvOk {
		return
	}

	e.Player.Move(pv)
}

func (e *Engine) AlignPlayerY() {
	var pv geometry.Vector
	var pvOk bool
	var collision object.Collidable

	extendedRect := e.Player.Rectangle()
	extendedRect.BottomY += 1e-12

	for t := range Collide2(extendedRect, e.Platforms, e.Tiles) {
		pv, pvOk = t.Rectangle().PushVectorY(e.Player.Rectangle()), true
		collision = t
		break
	}

	if !pvOk {
		// No collision -> in the air.
		e.Player.SetOnGround(nil, e.Tick)
		e.Player.Acceleration.Y = physics.GravityAcceleration
		return
	}

	if pv.Y <= 0 {
		// Zero can only be on ground since we extended only BottomY.
		e.Player.SetOnGround(collision, e.Tick)
		e.Player.Acceleration.Y = 0
	} else {
		// Collision with top -> in the air.
		e.Player.SetOnGround(nil, e.Tick)
		e.Player.Acceleration.Y = physics.GravityAcceleration
	}

	e.Player.Move(pv)

	if collision != e.Player.PrevGround() {
		// Negative force when we hit a new ground.
		e.Player.Speed.Y = 0
		if moving, ok := collision.(physics.Moving); ok {
			e.Player.Acceleration.Y += moving.SpeedVec().Y
		}
	}
}

func (e *Engine) CollectItems() error {
	collectedSomething := false

	for it := range Collide(e.Player.Rectangle(), e.Items) {
		if it.Collected {
			continue
		}

		e.Player.Collect(it)

		collectedSomething = true
	}

	if collectedSomething {
		snapshot := e.MakeSnapshot()
		if err := e.SaveSnapshot(snapshot); err != nil {
			return fmt.Errorf("saving snapshot: %w", err)
		}
	}

	return nil
}

func (e *Engine) CheckPortals() {
	for p := range Collide(e.Player.Rectangle(), e.Portals) {
		dx := 32.0
		if e.Player.Speed.X < 0 {
			e.Player.MoveTo(p.TeleportTo.Add(geometry.Vector{
				X: -dx,
				Y: 0,
			}))
		} else {
			e.Player.MoveTo(p.TeleportTo.Add(geometry.Vector{
				X: dx,
				Y: 0,
			}))
		}

		if p.Boss != "" {
			e.BossEntered = true
			e.BossPortal.MoveTo(geometry.Point{X: -9999, Y: -9999})
		}
	}
}

func (e *Engine) CheckSpikes() {
	for s := range Collide(e.Player.Rectangle(), e.Spikes) {
		e.Player.Health -= s.Damage
	}
}

func (e *Engine) CheckEnemyBullets() {
	var bullets []*damage.Bullet

	const maxBullets = 1000
	if len(e.EnemyBullets) > maxBullets {
		e.EnemyBullets = e.EnemyBullets[len(e.EnemyBullets)-maxBullets:]
	}

	rnd := rand.New(rand.NewPCG(0, uint64(e.Tick)))
	for _, b := range e.EnemyBullets {
		if b.PlayerSeekSpeed > 0 {
			b.Direction = e.Player.Rectangle().Center().SubPoint(b.Origin).Normalize().Multiply(b.PlayerSeekSpeed)
			dx := (rnd.Float64() - 0.5) / 3
			dy := (rnd.Float64() - 0.5) / 3
			b.Direction = b.Direction.Add(geometry.Vector{X: dx, Y: dy})
		}
		b.Move(b.Direction)

		ok := true
		for range Collide2(b.Rectangle(), e.Tiles, e.Platforms) {
			ok = false
			break
		}
		if ok {
			bullets = append(bullets, b)
		}
	}

	e.EnemyBullets = bullets

	for b := range Collide(e.Player.Rectangle(), e.EnemyBullets) {
		if b.Triggered {
			continue
		}

		e.Player.Health -= b.Damage
		b.Triggered = true
	}
}

func (e *Engine) CheckBoss() {
	if e.Boss == nil || !e.BossEntered {
		return
	}

	res := e.Boss.Tick(&boss.TickState{CurrentTick: e.Tick})

	if res.Dead {
		e.BossItem.MoveTo(*e.BossWinPoint)
		e.BossPortal.MoveTo(e.BossWinPoint.Add(geometry.Vector{X: -e.BossPortal.Width, Y: 0}))
	}

	e.EnemyBullets = append(e.EnemyBullets, res.Bullets...)
}

func (e *Engine) CheckNPCClose() *npc.NPC {
	for n := range Collide(e.Player.Rectangle().Extended(40), e.NPCs) {
		return n
	}

	return nil
}

func (e *Engine) CheckArcadeClose() *arcade.Machine {
	for a := range Collide(e.Player.Rectangle().Extended(40), e.Arcades) {
		return a
	}

	return nil
}

func (e *Engine) CheckSlotsClose() *casino.SlotMachine {
	for s := range Collide(e.Player.Rectangle(), e.Slots) {
		if s.LastTriggered+s.Cooldown > e.Tick {
			continue
		}

		return s
	}

	return nil
}

func (e *Engine) PlaySlots(s *casino.SlotMachine) {
	if s.Won() {
		e.Player.Coins += s.Coins
		e.notificationText = "Congrats! You won " + strconv.Itoa(s.Coins) + " coins!"
	} else {
		e.notificationText = "Aww. Try again next time!"
	}
	e.notificationEndTick = e.Tick + ebiten.TPS()*3

	s.LastTriggered = e.Tick
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

func (e *Engine) ActiveArcade() *arcade.Machine {
	return e.activeArcade
}
