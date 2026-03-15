package main

import (
	"fmt"

	"github.com/leonlau/xiangqi"
	xq "github.com/leonlau/xiangqi-interface"
)

func parsePosition(fen string) (*xiangqi.Position, error) {
	var pos xiangqi.Position
	if err := pos.UnmarshalText([]byte(fen)); err != nil {
		return nil, err
	}
	return &pos, nil
}

func convertMove(m *xiangqi.Move) xq.Move {
	s1 := m.S1()
	s2 := m.S2()
	var tags xq.MoveTag
	if m.HasTag(xiangqi.Capture) {
		tags |= xq.TagCapture
	}
	if m.HasTag(xiangqi.Check) {
		tags |= xq.TagCheck
	}
	return xq.Move{
		S1:   xq.Square{File: xq.File(s1.File()), Rank: xq.Rank(s1.Rank())},
		S2:   xq.Square{File: xq.File(s2.File()), Rank: xq.Rank(s2.Rank())},
		Tags: tags,
	}
}

// formatUCI 把 xq.Square 编码为 UCI 字符。xiangqi 的 UCINotation.Decode
// 期望 rank 是 1-10(1-indexed),所以这里 +1。
func formatUCI(s xq.Square) string {
	return fmt.Sprintf("%s%d", s.File.String(), int(s.Rank)+1)
}

// flattenComments 把 xiangqi 的 [][]string(每步多条评注)压成 []string(每步第一条)。
func flattenComments(in [][]string) []string {
	out := make([]string, len(in))
	for i, c := range in {
		if len(c) > 0 {
			out[i] = c[0]
		}
	}
	return out
}
