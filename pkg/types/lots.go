package types

type Lot struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       Price  `json:"price"`
	Seller      Seller `json:"seller"`
}
