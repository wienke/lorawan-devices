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
	"image/png"
	"testing"
)

func canvas(size int, withPill bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	if withPill {
		draw.Draw(img, image.Rect(744, 1092, 1164, 1095), image.NewUniform(pillBorder), image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(744, 1161, 1164, 1164), image.NewUniform(pillBorder), image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(900, 1120, 1000, 1140), image.NewUniform(color.Black), image.Point{}, draw.Src) // "text"
	}
	return img
}

func encode(t *testing.T, img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestStripPill(t *testing.T) {
	in := encode(t, canvas(pillCanvas, true))
	out, _, err := image.Decode(bytes.NewReader(stripPill(in, "device.png")))
	if err != nil {
		t.Fatal(err)
	}
	for y := pillBox.Min.Y; y < pillBox.Max.Y; y++ {
		for x := pillBox.Min.X; x < pillBox.Max.X; x++ {
			if !near(out.At(x, y), color.RGBA{255, 255, 255, 255}, 0) {
				t.Fatalf("pixel %d,%d not cleared", x, y)
			}
		}
	}
}

func TestStripPillLeavesOtherPhotos(t *testing.T) {
	for name, img := range map[string]image.Image{
		"no pill":    canvas(pillCanvas, false),
		"other size": canvas(1000, false),
	} {
		in := encode(t, img)
		if !bytes.Equal(stripPill(in, "device.png"), in) {
			t.Errorf("%s: photo was modified", name)
		}
	}
	if got := stripPill([]byte("not an image"), "x.png"); string(got) != "not an image" {
		t.Error("undecodable input was modified")
	}
}
