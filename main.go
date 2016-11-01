// This program implements the salt and pepper noise reduction algorithm.
// Additional sample images can be downloaded from
// http://www.fit.vutbr.cz/~vasicek/imagedb/?lev=75

package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
)

type bright uint32
type dirtyImg struct {
	img    image.Image
	salt   bright
	pepper bright
}

func calcBright(c color.Color) bright {
	r, g, b, _ := c.RGBA()
	return bright(r + g + b)
}

type colors []color.Color

func (colors colors) Len() int { return len(colors) }
func (colors colors) Less(i, j int) bool {
	ib := calcBright(colors[i])
	ij := calcBright(colors[j])
	return ib < ij
}
func (colors colors) Swap(i, j int) {
	colors[i], colors[j] = colors[j], colors[i]
}

func getColor(x, y int, img image.Image) (c color.Color, ok bool) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return color.Black, false
	}
	return img.At(x, y), true
}

func (di dirtyImg) collectColor(x, y int, colors *colors) {
	c, ok := getColor(x, y, di.img)
	if ok && !di.isDirty(c) {
		*colors = append(*colors, c)
	}
}

func (di dirtyImg) collectRing(x, y int, colors *colors, ring int) {
	for i := x - ring; i < x+ring; i++ {
		di.collectColor(i, y-ring, colors)
	}
	for i := x - ring; i < x+ring; i++ {
		di.collectColor(i, y+ring, colors)
	}
	for i := y - ring + 1; i < y+ring-1; i++ {
		di.collectColor(x-ring, i, colors)
	}
	for i := y - ring + 1; i < y+ring-1; i++ {
		di.collectColor(x+ring, i, colors)
	}
}

func (di dirtyImg) calculateColor(x, y int) color.Color {
	var colors colors
	for ring := 1; len(colors) == 0; ring++ {
		di.collectRing(x, y, &colors, ring)
	}
	sort.Sort(colors)
	return colors[len(colors)/2]
}

func (di dirtyImg) isDirty(c color.Color) bool {
	b := calcBright(c)
	return b == di.salt || b == di.pepper
}

func (di *dirtyImg) findSaltAndPepper() {
	bounds := di.img.Bounds()
	di.salt = 0
	di.pepper = math.MaxUint32
	for y := bounds.Min.X; y < bounds.Max.X; y++ {
		for x := bounds.Min.Y; x < bounds.Max.Y; x++ {
			b := calcBright(di.img.At(x, y))
			if di.salt < b {
				di.salt = b
			}
			if di.pepper > b {
				di.pepper = b
			}
		}
	}
}

func (di *dirtyImg) cleanUp() {
	di.findSaltAndPepper()

	bounds := di.img.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := di.img.At(x, y)
			if di.isDirty(c) {
				c = di.calculateColor(x, y)
			}
			dst.Set(x, y, c)
		}
	}
	di.img = dst
}

func main() {
	if len(os.Args) < 3 {
		panic("Input and output filenames are missing.")
	}

	in, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer in.Close()

	img, err := png.Decode(in)
	if err != nil {
		panic(err)
	}

	di := dirtyImg{img: img}
	di.cleanUp()

	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer out.Close()
	err = png.Encode(out, di.img)
	if err != nil {
		panic(err)
	}
}
