package shopify

import (
	"context"
	"net/http"
	"time"
)

type ProductService service

type Product struct {
	ID             int64       `json:"id"`
	Title          string      `json:"title"`
	BodyHTML       string      `json:"body_html"`
	Vendor         string      `json:"vendor"`
	ProductType    string      `json:"product_type"`
	Handle         string      `json:"handle"`
	TemplateSuffix interface{} `json:"template_suffix"`
	PublishedScope string      `json:"published_scope"`
	Tags           string      `json:"tags"`

	Variants []*ProductVariant `json:"variants"`
	Options  []*ProductOption  `json:"options"`
	Images   []*ProductImage   `json:"images"`
	Image    *ProductImage     `json:"image"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`
}

type ProductVariant struct {
	ID                   int64       `json:"id"`
	ProductID            int64       `json:"product_id"`
	Title                string      `json:"title"`
	Price                string      `json:"price"`
	Sku                  string      `json:"sku"`
	Position             int         `json:"position"`
	Grams                int         `json:"grams"`
	InventoryPolicy      string      `json:"inventory_policy"`
	FulfillmentService   string      `json:"fulfillment_service"`
	InventoryManagement  string      `json:"inventory_management"`
	Option1              string      `json:"option1"`
	Option2              string      `json:"option2"`
	Option3              string      `json:"option3"`
	Taxable              bool        `json:"taxable"`
	Barcode              string      `json:"barcode"`
	ImageID              interface{} `json:"image_id"`
	CompareAtPrice       interface{} `json:"compare_at_price"`
	InventoryQuantity    int         `json:"inventory_quantity"`
	Weight               float64     `json:"weight"`
	WeightUnit           string      `json:"weight_unit"`
	OldInventoryQuantity int         `json:"old_inventory_quantity"`
	RequiresShipping     bool        `json:"requires_shipping"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductOption struct {
	ID        int64    `json:"id"`
	ProductID int64    `json:"product_id"`
	Name      string   `json:"name"`
	Position  int      `json:"position"`
	Values    []string `json:"values"`
}
type ProductImages struct {
	ID         int64         `json:"id"`
	ProductID  int64         `json:"product_id"`
	Position   int           `json:"position"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
	Src        string        `json:"src"`
	VariantIds []interface{} `json:"variant_ids"`
}

type ProductImage struct {
	ID         int64         `json:"id"`
	ProductID  int64         `json:"product_id"`
	Position   int           `json:"position"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
	Src        string        `json:"src"`
	VariantIds []interface{} `json:"variant_ids"`
}

func (p *ProductService) List(ctx context.Context) ([]*Product, *http.Response, error) {
	req, err := p.client.NewRequest("GET", "/admin/products.json", nil)
	if err != nil {
		return nil, nil, err
	}

	var productWrapper struct {
		Products []*Product `json:"products"`
	}
	resp, err := p.client.Do(ctx, req, &productWrapper)
	if err != nil {
		return nil, resp, err
	}

	return productWrapper.Products, resp, nil
}