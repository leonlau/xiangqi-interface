package main

import (
	"github.com/leonlau/xiangqi/zobrist"
)

func (XiangqiEngineImpl) ZobristHash(fen string) (uint64, error) {
	return zobrist.Hash(fen)
}
