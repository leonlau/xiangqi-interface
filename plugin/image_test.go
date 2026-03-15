package main

import (
	"strings"
	"testing"

	xq "github.com/leonlau/xiangqi-interface"
)

const imageTestFEN = "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

func TestImageSVG(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	h, err := impl.NewImageEncoder()
	if err != nil {
		t.Fatalf("NewImageEncoder: %v", err)
	}
	defer impl.CloseImageEncoder(h)

	got, err := impl.ImageSVG(h, imageTestFEN)
	if err != nil {
		t.Fatalf("ImageSVG: %v", err)
	}
	if !strings.Contains(got, "<svg") {
		t.Fatalf("SVG output missing <svg>: %s", got[:min(200, len(got))])
	}
}

func TestImageSVGWithOptions(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	h, _ := impl.NewImageEncoder()
	defer impl.CloseImageEncoder(h)

	got, err := impl.ImageSVG(h, imageTestFEN,
		xq.ImagePerspectiveOpt{xq.Red},
		xq.ImageSquareColorsOpt{
			Light: xq.ImageColor{R: 240, G: 220, B: 180, A: 255},
			Dark:  xq.ImageColor{R: 180, G: 140, B: 100, A: 255},
		},
		xq.ImageMarkSquaresOpt{
			Color:   xq.ImageColor{R: 255, G: 0, B: 0, A: 200},
			Squares: []xq.Square{{File: xq.FileE, Rank: xq.Rank4}},
		},
		xq.ImageMarkArrowsOpt{
			Arrows: []xq.ImageArrow{
				{
					From:  xq.Square{File: xq.FileE, Rank: xq.Rank0},
					To:    xq.Square{File: xq.FileE, Rank: xq.Rank9},
					Color: xq.ImageColor{R: 0, G: 0, B: 255, A: 200},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("ImageSVG: %v", err)
	}
	if !strings.Contains(got, "<svg") {
		t.Fatalf("SVG output missing <svg>")
	}
}

func TestImageInvalidHandle(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	if _, err := impl.ImageSVG(xq.InvalidImageEncoderHandle, imageTestFEN); err == nil {
		t.Fatal("expected error for invalid handle")
	}
	if err := impl.CloseImageEncoder(xq.ImageEncoderHandle(99999)); err == nil {
		t.Fatal("expected error closing nonexistent handle")
	}
}
