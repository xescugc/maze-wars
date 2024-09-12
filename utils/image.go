package utils

import "image"

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}
