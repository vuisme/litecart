package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/vuisme/litecart/internal/queries"
	"github.com/vuisme/litecart/pkg/logging"
	"github.com/vuisme/litecart/pkg/webutil"
)

// Products is ...
// [get] /api/products
func Products(c *fiber.Ctx) error {
	db := queries.DB()
	log := logging.New()

	products, err := db.ListProducts(c.Context(), false)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Products", products)
}

// GetProduct is ...
// [get] /api/products/:product_id
func Product(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	product, err := db.Product(c.Context(), false, productID)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product info", product)
}
