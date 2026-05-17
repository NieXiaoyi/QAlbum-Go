package phash

import (
	"image"
	"math"
)

const (
	hashSize = 8
	imgSize  = 32
)

type FileWithPHash struct {
	ID    int
	PHash uint64
	File  interface{}
}

type Group struct {
	Files      []*FileWithPHash
	Similarity int
}

func ComputeDCT(img image.Image) uint64 {
	resized := resize(img, imgSize, imgSize)
	grayscale := toGrayscale(resized)
	dct := compute2DDCT(grayscale)
	return extractHash(dct)
}

func resize(img image.Image, width, height int) [][]float64 {
	srcBounds := img.Bounds()
	dst := make([][]float64, height)
	for y := 0; y < height; y++ {
		dst[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			srcX := x * srcBounds.Dx() / width
			srcY := y * srcBounds.Dy() / height
			r, g, b, _ := img.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY).RGBA()
			dst[y][x] = float64(r+g+b) / 3.0 / 65535.0
		}
	}
	return dst
}

func toGrayscale(data [][]float64) [][]float64 {
	size := len(data)
	result := make([][]float64, size)
	for i := 0; i < size; i++ {
		result[i] = make([]float64, size)
		for j := 0; j < size; j++ {
			result[i][j] = data[i][j]
		}
	}
	return result
}

func compute2DDCT(data [][]float64) [][]float64 {
	n := len(data)
	dct := make([][]float64, n)
	for i := 0; i < n; i++ {
		dct[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			sum := 0.0
			for x := 0; x < n; x++ {
				for y := 0; y < n; y++ {
					sum += data[x][y] * math.Cos((2*float64(x)+1)*math.Pi*float64(i)/(2*float64(n))) *
						math.Cos((2*float64(y)+1)*math.Pi*float64(j)/(2*float64(n)))
				}
			}
			ci := 1.0
			cj := 1.0
			if i == 0 {
				ci = 1.0 / math.Sqrt(2)
			}
			if j == 0 {
				cj = 1.0 / math.Sqrt(2)
			}
			dct[i][j] = (2.0 / float64(n)) * ci * cj * sum
		}
	}
	return dct
}

func extractHash(dct [][]float64) uint64 {
	avg := dct[0][0]
	var hash uint64
	for i := 1; i <= hashSize; i++ {
		for j := 1; j <= hashSize; j++ {
			hash <<= 1
			if dct[i][j] > avg {
				hash |= 1
			}
		}
	}
	return hash
}

func HammingDistance(a, b uint64) int {
	diff := a ^ b
	distance := 0
	for diff != 0 {
		distance++
		diff &= diff - 1
	}
	return distance
}

func ClusterByHammingDistance(files []*FileWithPHash, threshold int) []*Group {
	if len(files) == 0 {
		return nil
	}

	visited := make(map[int]bool)
	var groups []*Group

	for i, file := range files {
		if visited[i] {
			continue
		}

		var currentGroup []*FileWithPHash
		maxDistance := 0

		for j, other := range files {
			if i == j {
				continue
			}

			distance := HammingDistance(file.PHash, other.PHash)
			if distance <= threshold {
				visited[j] = true
				currentGroup = append(currentGroup, other)
				if distance > maxDistance {
					maxDistance = distance
				}
			}
		}

		if len(currentGroup) > 0 {
			currentGroup = append([]*FileWithPHash{file}, currentGroup...)
			visited[i] = true
			groups = append(groups, &Group{
				Files:      currentGroup,
				Similarity: maxDistance,
			})
		}
	}

	return groups
}
