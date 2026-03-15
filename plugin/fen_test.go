package main

import "testing"

func TestValidateFEN_Valid(t *testing.T) {
	engine := XiangqiEngineImpl{}
	fen := "rnbakabnr/9/1c5c1/p1p1p1p1p/9/9/P1P1P1P1P/1C5C1/9/RNBAKABNR w - - 0 1"
	if err := engine.ValidateFEN(fen); err != nil {
		t.Errorf("ValidateFEN(%q) = %v, want nil", fen, err)
	}
}

func TestValidateFEN_Invalid(t *testing.T) {
	engine := XiangqiEngineImpl{}
	cases := []string{
		"",
		"not a fen string",
	}
	for _, fen := range cases {
		if err := engine.ValidateFEN(fen); err == nil {
			t.Errorf("ValidateFEN(%q) = nil, want error", fen)
		}
	}
}
