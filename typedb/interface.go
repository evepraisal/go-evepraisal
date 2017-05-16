package typedb

type TypeDB interface {
	GetType(typeName string) (EveType, bool)
	HasType(typeName string) bool
	Close() error
}

type EveType struct {
	ID     int64   `json:"id"`
	Name   string  `json:"name"`
	Volume float64 `json:"volume"`
}
