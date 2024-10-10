package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/vuisme/litecart/internal/models"
	"github.com/vuisme/litecart/internal/queries"
	"github.com/vuisme/litecart/pkg/logging"
	"github.com/vuisme/litecart/pkg/webutil"
)

// Install is ...
// [post] /api/install
func Install(c *fiber.Ctx) error {
	db := queries.DB()
	log := logging.New()
	request := new(models.Install)

	if err := c.BodyParser(request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	if err := request.Validate(); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	if err := db.Install(c.Context(), request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Cart installed", nil)
}
