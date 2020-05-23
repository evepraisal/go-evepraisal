package typedb

// TypeDB defines the interface for storing and retreiving types
type TypeDB interface {
	GetType(typeName string) (EveType, bool)
	HasType(typeName string) bool
	GetTypeByID(typeID int64) (EveType, bool)
	ListTypes(startingTypeID int64, limit int64) ([]EveType, error)
	PutTypes([]EveType) error
	Search(s string) []EveType
	Delete() error
	Close() error
}

// EveType holds the information on a single type
type EveType struct {
	ID                int64       `json:"id"`
	GroupID           int64       `json:"group_id"`
	MarketGroupID     int64       `json:"market_group_id"`
	Name              string      `json:"name"`
	Aliases           []string    `json:"aliases"`
	Volume            float64     `json:"volume"`
	PackagedVolume    float64     `json:"packaged_volume"`
	BasePrice         float64     `json:"base_price"`
	BlueprintProducts []Component `json:"blueprint_products,omitempty"`
	Components        []Component `json:"components,omitempty"`
	BaseComponents    []Component `json:"base_components,omitempty"`
}

// Component defines what is needed and how many is needed to make something else
type Component struct {
	Quantity int64 `json:"quantity"`
	TypeID   int64 `json:"type_id"`
}
