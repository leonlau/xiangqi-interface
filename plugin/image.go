package main

import (
	"bytes"
	"fmt"
	"image/color"
	"sync"

	"github.com/leonlau/xiangqi"
	xq "github.com/leonlau/xiangqi-interface"
	xiangqiimage "github.com/leonlau/xiangqi/image"
)

var (
	imageMu       sync.Mutex
	imageCounter  int64
	imageEncoders = map[xq.ImageEncoderHandle]*xiangqiimage.Encoder{}
)

func getImageEncoder(h xq.ImageEncoderHandle) (*xiangqiimage.Encoder, error) {
	imageMu.Lock()
	defer imageMu.Unlock()
	e, ok := imageEncoders[h]
	if !ok {
		return nil, fmt.Errorf("invalid image encoder handle %d", h)
	}
	return e, nil
}

func imageColorToStd(c xq.ImageColor) color.Color {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// imageArrowWrapper 持有 from/to/color,通过回调构造 xiangqi 私有 arrow。
type imageArrowWrapper struct {
	from  xiangqi.Square
	to    xiangqi.Square
	color color.Color
}

// imageArrowFactoryFunc 是 xiangqiimage.MarkArrows 接受的 func 类型,这里不可见。
// 解决方案:让 xiangqiimage 包导出 Arrow(from, to).WithColor 后通过 xiangqiimage.MarkArrows 直接传入,
// 但 arrow 是私有类型。最终我们在 plugin 内部用 xiangqiimage.Arrow + .WithColor 链式调用,但不能直接传 MarkArrows。
//
// 改:在 plugin 内部直接调用 xiangqiimage.SVG 时,opts 列表里允许包含一个"构造箭头并加入 Encoder"的闭包。
// 为此我们在 xiangqi/image 包没有公开入口的情况下,把 MarkArrowsOpt 转写为 Encoder 私有 hook。

func applyImageOptions(opts []xq.ImageOption, arrows *[]imageArrowWrapper) []func(*xiangqiimage.Encoder) {
	var out []func(*xiangqiimage.Encoder)
	for _, o := range opts {
		switch v := o.(type) {
		case xq.ImagePerspectiveOpt:
			out = append(out, xiangqiimage.Perspective(xiangqi.Color(v.Perspective)))
		case xq.ImageSquareColorsOpt:
			out = append(out, xiangqiimage.SquareColors(imageColorToStd(v.Light), imageColorToStd(v.Dark)))
		case xq.ImageMarkSquaresOpt:
			sqs := make([]xiangqi.Square, len(v.Squares))
			for i, sq := range v.Squares {
				sqs[i] = xiangqi.NewSquare(xiangqi.File(sq.File), xiangqi.Rank(sq.Rank))
			}
			out = append(out, xiangqiimage.MarkSquares(imageColorToStd(v.Color), sqs...))
		case xq.ImageMarkArrowsOpt:
			for _, a := range v.Arrows {
				*arrows = append(*arrows, imageArrowWrapper{
					from:  xiangqi.NewSquare(xiangqi.File(a.From.File), xiangqi.Rank(a.From.Rank)),
					to:    xiangqi.NewSquare(xiangqi.File(a.To.File), xiangqi.Rank(a.To.Rank)),
					color: imageColorToStd(a.Color),
				})
			}
			// 把箭头追加到一个 closure,稍后在 ImageSVG 里通过 xiangqiimage.MarkArrows 注入。
			captured := *arrows
			out = append(out, func(e *xiangqiimage.Encoder) {
				for _, w := range captured {
					xiangqiimage.MarkArrows(xiangqiimage.Arrow(w.from, w.to).WithColor(w.color))(e)
				}
			})
		}
	}
	return out
}

func (XiangqiEngineImpl) NewImageEncoder() (xq.ImageEncoderHandle, error) {
	imageMu.Lock()
	imageCounter++
	h := xq.ImageEncoderHandle(imageCounter)
	imageEncoders[h] = &xiangqiimage.Encoder{}
	imageMu.Unlock()
	return h, nil
}

func (XiangqiEngineImpl) CloseImageEncoder(h xq.ImageEncoderHandle) error {
	imageMu.Lock()
	_, ok := imageEncoders[h]
	if ok {
		delete(imageEncoders, h)
	}
	imageMu.Unlock()
	if !ok {
		return fmt.Errorf("invalid image encoder handle %d", h)
	}
	return nil
}

func (XiangqiEngineImpl) ImageSVG(h xq.ImageEncoderHandle, fen string, opts ...xq.ImageOption) (string, error) {
	if _, err := getImageEncoder(h); err != nil {
		return "", err
	}
	pos, err := parsePosition(fen)
	if err != nil {
		return "", err
	}
	var arrows []imageArrowWrapper
	xiangqiOpts := applyImageOptions(opts, &arrows)
	var buf bytes.Buffer
	if err := xiangqiimage.SVG(&buf, pos.Board(), xiangqiOpts...); err != nil {
		return "", fmt.Errorf("render svg: %w", err)
	}
	return buf.String(), nil
}
