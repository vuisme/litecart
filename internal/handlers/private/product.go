package handlers

import (
	"fmt"

	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/vuisme/litecart/internal/models"
	"github.com/vuisme/litecart/internal/queries"
	"github.com/vuisme/litecart/pkg/errors"
	"github.com/vuisme/litecart/pkg/fsutil"
	"github.com/vuisme/litecart/pkg/logging"
	"github.com/vuisme/litecart/pkg/webutil"
)

// Products is ...
// [get] /api/_/products
func Products(c *fiber.Ctx) error {
	db := queries.DB()
	log := logging.New()

	products, err := db.ListProducts(c.Context(), true)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Products", products)
}

// AddProduct is ...
// [post] /api/_/products
func AddProduct(c *fiber.Ctx) error {
	db := queries.DB()
	log := logging.New()
	request := &models.Product{}

	if err := c.BodyParser(request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	product, err := db.AddProduct(c.Context(), request)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product added", product)
}

// GetProduct is ...
// [get] /api/_/products/:product_id
func Product(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	product, err := db.Product(c.Context(), true, productID)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product info", product)
}

// UpdateProduct is ...
// [patch] /api/_/products/:product_id
func UpdateProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()
	request := new(models.Product)
	request.ID = productID

	if err := c.BodyParser(request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	if err := db.UpdateProduct(c.Context(), request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product updated", nil)
}

// DeleteProduct is ...
// [delete] /api/_/products/:product_id
func DeleteProduct(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	if err := db.DeleteProduct(c.Context(), productID); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product deleted", nil)
}

// UpdateProductActive is ...
// [patch] /api/_/products/:product_id/active
func UpdateProductActive(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	if err := db.UpdateActive(c.Context(), productID); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product active updated", nil)
}

// ProductImages
// [get] /api/_/products/:product_id/image
func ProductImages(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	images, err := db.ProductImages(c.Context(), productID)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product images", images)
}

// AddProductImage is ...
// [post] /api/_/products/:product_id/image
func AddProductImage(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	file, err := c.FormFile("document")
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	validMIME := false
	validMIMETypes := []string{"image/png", "image/jpeg"}
	for _, mime := range validMIMETypes {
		if mime == file.Header["Content-Type"][0] {
			validMIME = true
		}
	}
	if !validMIME {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, "file format not supported")
	}

	fileUUID := uuid.New().String()
	fileExt := fsutil.ExtName(file.Filename)
	fileName := fmt.Sprintf("%s.%s", fileUUID, fileExt)
	filePath := fmt.Sprintf("./lc_uploads/%s", fileName)
	fileOrigName := file.Filename

	c.SaveFile(file, filePath)

	fileSource, err := imaging.Open(filePath)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	sizes := []struct {
		size string
		dim  int
	}{
		{"sm", 147},
		{"md", 400},
	}

	for _, s := range sizes {
		resizedImage := imaging.Fill(fileSource, s.dim, s.dim, imaging.Center, imaging.Lanczos)
		err := imaging.Save(resizedImage, fmt.Sprintf("./lc_uploads/%s_%s.%s", fileUUID, s.size, fileExt))
		if err != nil {
			log.ErrorStack(err)
			return webutil.StatusInternalServerError(c)
		}
	}

	addedImage, err := db.AddImage(c.Context(), productID, fileUUID, fileExt, fileOrigName)
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Image added", addedImage)
}

// DeleteProductImage is ...
// [delete] /api/_/products/:product_id/image/:image_id
func DeleteProductImage(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	imageID := c.Params("image_id")
	db := queries.DB()
	log := logging.New()

	if err := db.DeleteImage(c.Context(), productID, imageID); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Image deleted", nil)
}

// ProductDigital
// [get] /api/_/products/:product_id/digital
func ProductDigital(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	digital, err := db.ProductDigital(c.Context(), productID)
	if err != nil {
		if err == errors.ErrProductNotFound {
			return webutil.StatusNotFound(c)
		}
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Product digital", digital)
}

// AddProductDigital is ...
// [post] /api/_/products/:product_id/digital
func AddProductDigital(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	db := queries.DB()
	log := logging.New()

	fileTmp, _ := c.FormFile("document")
	if fileTmp != nil {
		fileUUID := uuid.New().String()
		fileExt := fsutil.ExtName(fileTmp.Filename)
		fileName := fmt.Sprintf("%s.%s", fileUUID, fileExt)
		filePath := fmt.Sprintf("./lc_digitals/%s", fileName)
		fileOrigName := fileTmp.Filename

		c.SaveFile(fileTmp, filePath)

		file, err := db.AddDigitalFile(c.Context(), productID, fileUUID, fileExt, fileOrigName)
		if err != nil {
			log.ErrorStack(err)
			return webutil.StatusInternalServerError(c)
		}

		return webutil.Response(c, fiber.StatusOK, "Digital added", file)
	}

	data, err := db.AddDigitalData(c.Context(), productID, "")
	if err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Digital added", data)
}

// UpdateProductDigital is ...
// [patch] /api/_/products/:product_id/digital/:digital_id
func UpdateProductDigital(c *fiber.Ctx) error {
	request := new(models.Data)
	request.ID = c.Params("digital_id")
	// request.Content = c.Params("digital_id")
	db := queries.DB()
	log := logging.New()

	if err := c.BodyParser(request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusBadRequest(c, err.Error())
	}

	if err := db.UpdateDigital(c.Context(), request); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Digital updated", nil)
}

// DeleteProductDigital is ...
// [delete] /api/_/products/:product_id/digital/:digital_id
func DeleteProductDigital(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	digitalID := c.Params("digital_id")
	db := queries.DB()
	log := logging.New()

	if err := db.DeleteDigital(c.Context(), productID, digitalID); err != nil {
		log.ErrorStack(err)
		return webutil.StatusInternalServerError(c)
	}

	return webutil.Response(c, fiber.StatusOK, "Digital deleted", nil)
}
