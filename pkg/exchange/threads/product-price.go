package threads

type Price struct {
	PriceGuid  string `json:"price_guid"`
	Value      int64  `json:"value"`
	LastUpdate string `json:"last_update"`
}

type ProductPrice struct {
	EntityId    int64   `json:"entity_id"`
	ProductGuid string  `json:"product_guid"`
	Prices      []Price `json:"prices"`
}

func (th *Thread) ProductPrice() {

}
