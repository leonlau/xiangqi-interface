package xq

type Color int8

const (
	// NoColor represents no color
	NoColor Color = iota
	// Red represents the color red
	Red
	// Black represents the color black
	Black
)

// String implements the fmt.Stringer interface and returns
// the color's FEN compatible notation.
func (c Color) String() string {
	switch c {
	case Red:
		return "w" // fen中还是用 w 表示红色
	case Black:
		return "b"
	}
	return "-"
}

type Game interface{}
