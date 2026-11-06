package main

import (
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var sampleButton = image.Rect(520, 370, 740, 406)
var backButton = image.Rect(320, 370, 480, 406)
var endlessCheckbox = image.Rect(240, 365, 400, 389)
var completionBackButton = image.Rect(240, 240, 560, 280)

const (
	windowWidth  = 800
	windowHeight = 420
)

type note struct {
	name  string
	image *ebiten.Image
}

type game struct {
	note       note
	difficulty *difficulty
	endless    bool
	remaining  []string
	completed  bool
}

func newGame() *game {
	return &game{endless: true}
}

func (g *game) startSession(level difficulty) error {
	if !g.endless {
		paths, err := notePaths(noteAssets, "violin_and_bass_notes", level)
		if err != nil {
			return err
		}
		rand.Shuffle(len(paths), func(i, j int) { paths[i], paths[j] = paths[j], paths[i] })
		if len(paths) > 1 && paths[0] == filepath.ToSlash(filepath.Join("violin_and_bass_notes", g.note.name)) {
			j := 1 + rand.IntN(len(paths)-1)
			paths[0], paths[j] = paths[j], paths[0]
		}
		g.remaining = paths
	}
	g.completed = false
	return g.sampleNote(level)
}

func (g *game) sampleNote(level difficulty) error {
	var img image.Image
	var name string
	var err error
	if g.endless {
		img, name, err = loadRandomNote(noteAssets, "violin_and_bass_notes", level, g.note.name)
	} else {
		if len(g.remaining) == 0 {
			if g.note.image != nil {
				g.note.image.Deallocate()
				g.note.image = nil
			}
			g.completed = true
			return nil
		}
		img, name, err = loadNote(noteAssets, "violin_and_bass_notes", g.remaining[0])
	}
	if err != nil {
		return err
	}
	if !g.endless {
		g.remaining = g.remaining[1:]
	}
	next := note{name: name, image: ebiten.NewImageFromImage(img)}
	if g.note.image != nil {
		g.note.image.Deallocate()
	}
	g.note = next
	g.difficulty = &level
	return nil
}

func (g *game) showDifficultySelection() {
	if g.note.image != nil {
		g.note.image.Deallocate()
		g.note.image = nil
	}
	g.difficulty = nil
	g.remaining = nil
	g.completed = false
}

func difficultyButton(index int) image.Rectangle {
	y := 110 + index*50
	return image.Rect(240, y, 560, y+40)
}

func (g *game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	x, y := ebiten.CursorPosition()
	cursor := image.Pt(x, y)
	ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	if g.difficulty == nil {
		if cursor.In(endlessCheckbox) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.endless = !g.endless
				return nil
			}
		}
		for i, level := range difficulties {
			if cursor.In(difficultyButton(i)) {
				ebiten.SetCursorShape(ebiten.CursorShapePointer)
				if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
					return g.startSession(level)
				}
			}
		}
		return nil
	}
	if g.completed {
		if cursor.In(completionBackButton) {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.showDifficultySelection()
			}
		}
		return nil
	}
	if cursor.In(backButton) {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.showDifficultySelection()
			return nil
		}
	}
	hovered := cursor.In(sampleButton)
	if hovered {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		(hovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)) {
		return g.sampleNote(*g.difficulty)
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 30, G: 34, B: 44, A: 255})
	drawText(screen, "Notomat", 60, 30, 28)
	if g.difficulty == nil {
		drawText(screen, "Choose a difficulty to start", 240, 76, 16)
		for i, level := range difficulties {
			drawButton(screen, difficultyButton(i), level.name+" - "+level.rangeLabel)
		}
		vector.FillRect(screen, float32(endlessCheckbox.Min.X), float32(endlessCheckbox.Min.Y),
			20, 20, color.White, false)
		if g.endless {
			vector.FillRect(screen, float32(endlessCheckbox.Min.X+4), float32(endlessCheckbox.Min.Y+4),
				12, 12, color.RGBA{R: 52, G: 100, B: 170, A: 255}, false)
		}
		drawText(screen, "Endless Mode", float64(endlessCheckbox.Min.X+30), float64(endlessCheckbox.Min.Y), 16)
		return
	}
	drawText(screen, g.difficulty.name+" - "+g.difficulty.rangeLabel, 520, 40, 16)
	if g.completed {
		drawCenteredText(screen, "good job", image.Rect(240, 150, 560, 190), 28)
		drawCenteredText(screen, "You have seen every note in this difficulty.", image.Rect(60, 195, 740, 225), 16)
		drawButton(screen, completionBackButton, "Back to difficulty selection")
		return
	}

	const panelX, panelY, panelWidth, panelHeight = 60, 90, 680, 240
	vector.FillRect(screen, panelX, panelY, panelWidth, panelHeight, color.White, false)
	bounds := g.note.image.Bounds()
	scale := math.Min(float64(panelWidth-20)/float64(bounds.Dx()), float64(panelHeight-20)/float64(bounds.Dy()))
	options := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	options.GeoM.Scale(scale, scale)
	options.GeoM.Translate(
		panelX+(panelWidth-float64(bounds.Dx())*scale)/2,
		panelY+(panelHeight-float64(bounds.Dy())*scale)/2,
	)
	screen.DrawImage(g.note.image, options)
	drawButton(screen, backButton, "Back")
	drawButton(screen, sampleButton, "New image [Space]")
}

func drawButton(screen *ebiten.Image, bounds image.Rectangle, label string) {
	buttonColor := color.RGBA{R: 52, G: 100, B: 170, A: 255}
	x, y := ebiten.CursorPosition()
	if image.Pt(x, y).In(bounds) {
		buttonColor = color.RGBA{R: 68, G: 122, B: 198, A: 255}
	}
	vector.FillRect(screen, float32(bounds.Min.X), float32(bounds.Min.Y),
		float32(bounds.Dx()), float32(bounds.Dy()), buttonColor, false)
	drawCenteredText(screen, label, bounds, 16)
}

func (g *game) Layout(_, _ int) (int, int) {
	return windowWidth, windowHeight
}

func main() {
	if fontError != nil {
		reportError(fontError)
		return
	}
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("Notomat")
	if err := ebiten.RunGame(newGame()); err != nil {
		reportError(err)
	}
}
