// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugo

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"strings"
)

// Harmonized device photos in the repository carry a "Works with The Things
// Stack" pill in a fixed spot in the bottom-right corner of a 1200x1200
// canvas. That spot is pure white before the pill is drawn, so the website
// shows the clean photo by painting the pill box white again.
//
// The geometry and border colour must match the photo harmonizer.
const pillCanvas = 1200

var (
	pillBox    = image.Rect(740, 1088, 1169, 1169) // cleared area, incl. antialias halo
	pillBorder = color.RGBA{200, 209, 221, 255}    // 3 px outline, #C8D1DD
	// Points on the pill outline, and points just outside it that must be white.
	pillBorderPoints  = []image.Point{{954, 1093}, {954, 1162}, {850, 1093}}
	pillOutsidePoints = []image.Point{{954, 1078}, {728, 1128}}
)

func near(c color.Color, want color.RGBA, tol int) bool {
	r, g, b, _ := c.RGBA()
	d := func(v uint32, w uint8) bool {
		x := int(v>>8) - int(w)
		return x >= -tol && x <= tol
	}
	return d(r, want.R) && d(g, want.G) && d(b, want.B)
}

func hasPill(img image.Image) bool {
	b := img.Bounds()
	if b.Dx() != pillCanvas || b.Dy() != pillCanvas {
		return false
	}
	for _, p := range pillBorderPoints {
		if !near(img.At(b.Min.X+p.X, b.Min.Y+p.Y), pillBorder, 12) {
			return false
		}
	}
	white := color.RGBA{255, 255, 255, 255}
	for _, p := range pillOutsidePoints {
		if !near(img.At(b.Min.X+p.X, b.Min.Y+p.Y), white, 10) {
			return false
		}
	}
	return true
}

// stripPill returns the photo without the pill. Photos without a pill (or
// that cannot be decoded) are returned unchanged.
func stripPill(input []byte, fileName string) []byte {
	img, format, err := image.Decode(bytes.NewReader(input))
	if err != nil || !hasPill(img) {
		return input
	}

	out := image.NewRGBA(img.Bounds())
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
	draw.Draw(out, pillBox.Add(out.Bounds().Min), image.NewUniform(color.White), image.Point{}, draw.Src)

	var buf bytes.Buffer
	ext := strings.ToLower(fileName)
	if format == "jpeg" || strings.HasSuffix(ext, ".jpg") || strings.HasSuffix(ext, ".jpeg") {
		err = jpeg.Encode(&buf, out, &jpeg.Options{Quality: 92})
	} else {
		err = png.Encode(&buf, out)
	}
	if err != nil {
		return input
	}
	return buf.Bytes()
}
