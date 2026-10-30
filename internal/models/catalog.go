package models

type ProductInput struct {
	Title       string   `json:"title" binding:"required,max=255"`
	Description string   `json:"description"`
	Price       *float64 `json:"price" binding:"required,gte=0,lte=9999999999"`
	Rating      *float64 `json:"rating" binding:"required,gte=0,lte=5"`
	Images      []string `json:"images" binding:"required,min=3"`
	Stock       *int     `json:"stock_quantity" binding:"required,gte=0"`
}

type ProductFilter struct {
	Page, Limit                   int
	Search, SortBy, Order         string
	MinPrice, MaxPrice, MinRating *float64
}

type ProductPage struct {
	Items []Product `json:"items"`
	Meta  PageMeta  `json:"meta"`
}

type PageMeta struct {
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
	CurrentPage int `json:"current_page"`
	Limit       int `json:"limit"`
}

type CartItem struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type Cart struct {
	Items              []CartItem `json:"items"`
	Subtotal           float64    `json:"subtotal"`
	DiscountPercentage int        `json:"discount_percentage"`
	Total              float64    `json:"total"`
}
