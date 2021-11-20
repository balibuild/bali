package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/balibuild/bali/v2/makeico"
)

// IconMaker icon maker
type IconMaker struct {
	images []image.Image
	fds    []*os.File
}

// Close close
func (m *IconMaker) Close() error {
	var err error
	for _, f := range m.fds {
		if e := f.Close(); e != nil {
			err = e
		}
	}
	return err
}

// AddImage add image path
func (m *IconMaker) AddImage(name string) error {
	fd, err := os.Open(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable open file: %v\n", err)
		return err
	}
	m.fds = append(m.fds, fd)
	img, err := png.Decode(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable decode png: %v\n", err)
		return err
	}
	m.images = append(m.images, img)
	return nil
}

// Make create icon
func (m *IconMaker) Make() error {
	if len(m.fds) == 0 {
		return errors.New("no image add")
	}
	name := m.fds[0].Name() + ".ico"
	fd, err := os.Create(name)
	if err != nil {
		return err
	}
	defer fd.Close()
	if err := makeico.EncodePNG(fd, m.images...); err != nil {
		return err
	}
	return nil
}

// .\ico.exe .\circular-chart.png .\circular-chart-256.png .\circular-chart-128.png .\circular-chart-64.png .\circular-chart-32.png .\circular-chart-24.png .\circular-chart-16.png

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s png\n", os.Args[0])
		os.Exit(1)
	}
	var m IconMaker
	defer m.Close()
	for i := 1; i < len(os.Args); i++ {
		m.AddImage(os.Args[i])
	}
	if err := m.Make(); err != nil {
		fmt.Fprintf(os.Stderr, "build icon error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "convert png to ico success\n")
}
