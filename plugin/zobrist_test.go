package main

import "testing"

func TestZobristHash(t *testing.T) {
	impl := &XiangqiEngineImpl{}
	startFEN := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"

	cases := []struct {
		name      string
		fen       string
		expectErr bool
	}{
		{"start red to move", startFEN, false},
		{"start black to move", "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR b - - 0 1", false},
		{"invalid fen", "garbage", true},
		{"empty fen", "", true},
	}
	seen := make(map[uint64]bool)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := impl.ZobristHash(c.fen)
			if c.expectErr {
				if err == nil {
					t.Fatalf("expected error, got %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ZobristHash: %v", err)
			}
			if got == 0 {
				t.Fatalf("got zero hash for %q", c.fen)
			}
			seen[got] = true
		})
	}

	// 不同走子方应得不同 hash
	hashRed, _ := impl.ZobristHash(startFEN)
	hashBlack, _ := impl.ZobristHash("rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR b - - 0 1")
	if hashRed == hashBlack {
		t.Fatalf("red/black hash should differ, both = %d", hashRed)
	}
}
