package typedb

type TypeDB interface {
	GetType(typeName string) (EveType, bool)
	HasType(typeName string) bool
	GetTypeByID(typeID int64) (EveType, bool)
	PutType(EveType) error
	Search(s string) []EveType
	Delete() error
	Close() error
}

type EveType struct {
	ID                int64       `json:"id"`
	GroupID           int64       `json:"group_id"`
	Name              string      `json:"name"`
	Volume            float64     `json:"volume"`
	BasePrice         float64     `json:"base_price"`
	BlueprintProducts []Component `json:"blueprint_products,omitempty"`
	BaseComponents    []Component `json:"components,omitempty"`
}

type Component struct {
	Quantity int64 `json:"quantity"`
	TypeID   int64 `json:"type_id"`
}
