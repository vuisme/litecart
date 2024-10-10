package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vuisme/litecart/internal/mailer"
	"github.com/vuisme/litecart/internal/queries"
	"github.com/vuisme/litecart/pkg/logging"
	"github.com/vuisme/litecart/pkg/webutil"
)

// Carts is ...
// [get] /api/_/carts
func Carts(c *fiber.Ctx) error {
	db := queries.DB()
	log := logging.New()

	products, err := db.Carts(c.Context())
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Carts", products)
}

// CartSendMail
// [post] /api/_/carts/:cart_id/mail
func CartSendMail(c *fiber.Ctx) error {
	cartID := c.Params("cart_id")
	log := logging.New()

	if err := mailer.SendCartLetter(cartID); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Mail sended", nil)
}
