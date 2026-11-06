package main

import (
	"bytes"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var uiFont, fontError = text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))

func drawText(screen *ebiten.Image, label string, x, y, size float64) {
	options := &text.DrawOptions{}
	options.GeoM.Translate(x, y)
	text.Draw(screen, label, &text.GoTextFace{Source: uiFont, Size: size}, options)
}

func drawCenteredText(screen *ebiten.Image, label string, bounds image.Rectangle, size float64) {
	face := &text.GoTextFace{Source: uiFont, Size: size}
	width, height := text.Measure(label, face, size)
	drawText(screen, label, float64(bounds.Min.X)+(float64(bounds.Dx())-width)/2,
		float64(bounds.Min.Y)+(float64(bounds.Dy())-height)/2, size)
}
