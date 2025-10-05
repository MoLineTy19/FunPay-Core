package types

type Lot struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       Price  `json:"price"`
	UserLink    string `json:"userLink"`
	UserName    string `json:"userName"`
}
