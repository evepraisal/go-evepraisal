package parsers

type ParserResult interface {
	Name() string
	Quantity() int64
	Volume() float64
}
