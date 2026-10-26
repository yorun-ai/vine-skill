package catalog

import (
	"go.yorun.ai/vine/core/ex"

	skeled "example.com/shop/skeled/catalog"
)

type CatalogService struct {
	skeled.DefaultCatalogServiceServer
}

func (*CatalogService) GetProduct(sku string) skeled.Product {
	if sku != "notebook" {
		ex.PanicNew(ex.NotFound, "product not found")
	}
	return skeled.Product{Sku: sku, Title: "Vine notebook", UnitPriceCents: 1290}
}
