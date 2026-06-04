package visual

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

// CompareImages compares two image files pixel by pixel.
// It returns a similarity score (0.0 to 1.0) and optionally writes a diff image to diffPath.
func CompareImages(baselinePath, currentPath, diffPath string) (float64, error) {
	bFile, err := os.Open(baselinePath)
	if err != nil {
		return 0, err
	}
	defer bFile.Close()

	cFile, err := os.Open(currentPath)
	if err != nil {
		return 0, err
	}
	defer cFile.Close()

	img1, _, err := image.Decode(bFile)
	if err != nil {
		return 0, err
	}
	img2, _, err := image.Decode(cFile)
	if err != nil {
		return 0, err
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	// Use the larger dimensions for the diff image
	maxX := bounds1.Max.X
	if bounds2.Max.X > maxX {
		maxX = bounds2.Max.X
	}
	maxY := bounds1.Max.Y
	if bounds2.Max.Y > maxY {
		maxY = bounds2.Max.Y
	}

	diffImg := image.NewRGBA(image.Rect(0, 0, maxX, maxY))
	diffCount := 0
	totalPixels := maxX * maxY

	for y := 0; y < maxY; y++ {
		for x := 0; x < maxX; x++ {
			var r1, g1, b1 uint32
			var r2, g2, b2 uint32

			if x < bounds1.Max.X && y < bounds1.Max.Y {
				r1, g1, b1, _ = img1.At(x, y).RGBA()
			}
			if x < bounds2.Max.X && y < bounds2.Max.Y {
				r2, g2, b2, _ = img2.At(x, y).RGBA()
			}

			// Simple Euclidean distance in RGB space
			dr := float64(r1>>8) - float64(r2>>8)
			dg := float64(g1>>8) - float64(g2>>8)
			db := float64(b1>>8) - float64(b2>>8)
			dist := math.Sqrt(dr*dr + dg*dg + db*db)

			if dist > 30.0 { // threshold for visual difference
				diffCount++
				diffImg.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red for diff
			} else {
				// Fade out the baseline image for context
				gray := uint8((float64(r1>>8) + float64(g1>>8) + float64(b1>>8)) / 3.0)
				faded := uint8(float64(gray)*0.2 + 200.0) // Lighten
				diffImg.Set(x, y, color.RGBA{faded, faded, faded, 255})
			}
		}
	}

	similarity := 1.0 - (float64(diffCount) / float64(totalPixels))

	// Save diff image
	if diffPath != "" {
		if err := os.MkdirAll(filepath.Dir(diffPath), 0755); err == nil {
			if out, err := os.Create(diffPath); err == nil {
				defer out.Close()
				png.Encode(out, diffImg)
			}
		}
	}

	return similarity, nil
}
