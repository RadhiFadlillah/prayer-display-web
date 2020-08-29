package backend

import (
	"fmt"
	"image"
	"image/color"
	"net"
	"os"
	"syscall"

	"github.com/lucasb-eyer/go-colorful"
	ce "github.com/marekm4/color-extractor"
)

func checkError(err error) {
	if err == nil {
		return
	}

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
				return
			}
		}
	}

	panic(err)
}

func getDominantColor(img image.Image) colorful.Color {
	colors := ce.ExtractColors(img)
	mainColor, ok := colorful.MakeColor(colors[0])
	if !ok {
		return colorful.Color{R: 0, G: 0, B: 0}
	}
	return mainColor
}

func getColorPalette(img image.Image) (main, accent, font colorful.Color) {
	// Get dominant color as main
	main = getDominantColor(img)

	// Get complementary color as accent
	h, _, l := main.Hsl()
	h -= 180
	if h < 0 {
		h += 360
	}

	if l >= 0.9 {
		l -= 0.2
	}

	if l <= 0.1 {
		l += 0.2
	}

	accent = colorful.Hsl(h, 1, l)

	// Get font color depending on lightness
	font = colorful.Color{R: 0, G: 0, B: 0}
	if l <= 0.5 {
		font, _ = colorful.MakeColor(color.White)
	}

	return
}

func colorToRGBA(color colorful.Color, alpha float64) string {
	r, g, b := color.RGB255()
	return fmt.Sprintf("rgba(%d, %d, %d, %.03f)", r, g, b, alpha)
}
