package makeico

// origin https://github.com/Kodeworks/golang-image-ico
// https://github.com/shibukawa/golang-image-ico

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	"image/png"
	"io"
)

type icondir struct {
	reserved  uint16
	imageType uint16
	numImages uint16
}

type icondirentry struct {
	imageWidth   uint8
	imageHeight  uint8
	numColors    uint8
	reserved     uint8
	colorPlanes  uint16
	bitsPerPixel uint16
	sizeInBytes  uint32
	offset       uint32
}

func newIcondirentry(offset int) icondirentry {
	var ide icondirentry
	ide.colorPlanes = 1         // windows is supposed to not mind 0 or 1, but other icon files seem to have 1 here
	ide.bitsPerPixel = 32       // can be 24 for bitmap or 24/32 for png. Set to 32 for now
	ide.offset = uint32(offset) //6 icondir + 16 icondirentry, next image will be this image size + 16 icondirentry, etc
	return ide
}

// EncodePNG encoding
func EncodePNG(w io.Writer, images ...image.Image) (err error) {
	id := icondir{
		imageType: 1,
		numImages: uint16(len(images)),
	}
	err = binary.Write(w, binary.LittleEndian, id)
	if err != nil {
		return
	}
	imageSizes := make([]int, len(images))

	pngbb := new(bytes.Buffer)
	for i, im := range images {
		prevSize := len(pngbb.Bytes())
		b := im.Bounds()
		m := image.NewRGBA(b)
		draw.Draw(m, b, im, b.Min, draw.Src)
		err = png.Encode(pngbb, m)
		if err != nil {
			return
		}
		imageSizes[i] = len(pngbb.Bytes()) - prevSize
	}
	offset := 6 + 16*len(images)
	for i, im := range images {
		bounds := im.Bounds()
		entry := icondirentry{
			imageWidth:   uint8(bounds.Dx()),
			imageHeight:  uint8(bounds.Dy()),
			colorPlanes:  1,
			bitsPerPixel: 32,
			sizeInBytes:  uint32(imageSizes[i]),
			offset:       uint32(offset),
		}
		offset += imageSizes[i]
		err = binary.Write(w, binary.LittleEndian, entry)
		if err != nil {
			return
		}
	}
	_, err = w.Write(pngbb.Bytes())
	return
}
