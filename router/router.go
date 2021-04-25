package router

import (
	"gr_addresses/service"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var app = fiber.New()

// InitRouter -
func InitRouter(port string) {
	Routes()
	app.Listen(port)
}

// Routes -
func Routes() {
	app.Use(recover.New())
	app.Get("/search", searchHandler)
}

// Search -
func searchHandler(c *fiber.Ctx) error {
	service := new(service.AddressSearch)

	// Filter by prefecture
	if c.Query("prefecture") != "" && c.Query("prefecture_id") == "" {
		prefectures, err := service.FilterPrefecture(c.Query("prefecture"))
		if err != nil {
			return internalServerError(err)
		}
		return c.JSON(fiber.Map{
			"results": prefectures,
		})
	}

	// Filter by city
	if c.Query("city") != "" && c.Query("city_id") == "" {
		cities, err := service.FilterCity(c.Query("city"))
		if err != nil {
			return internalServerError(err)
		}
		return c.JSON(fiber.Map{
			"results": cities,
		})
	}

	// Filter by zip code
	if c.Query("zipcode") != "" && c.Query("zipcode_id") == "" {
		zipcodes, err := service.FilterZipcode(c.Query("zipcode"))
		if err != nil {
			return internalServerError(err)
		}
		return c.JSON(fiber.Map{
			"results": zipcodes,
		})
	}

	// Find prefecture by id
	if c.Query("prefecture_id") != "" {
		var err error
		service, err = service.FindPrefectureByID(c.Query("prefecture_id"))
		if err != nil {
			return internalServerError(err)
		}
	}

	// Find city by id
	if c.Query("city_id") != "" {
		var err error
		service, err = service.FindCityByID(c.Query("city_id"))
		if err != nil {
			return internalServerError(err)
		}
	}

	// Find zipcode by id
	if c.Query("zipcode_id") != "" {
		var err error
		service, err = service.FindZipcodeByID(c.Query("zipcode_id"))
		if err != nil {
			return internalServerError(err)
		}
	}

	// Filter by street
	if c.Query("street") != "" {
		srv := service
		streets, err := srv.FilterStreet(c.Query("street"))
		if err != nil {
			return internalServerError(err)
		}
		return c.JSON(fiber.Map{
			"results": streets,
		})
	}

	return c.JSON(fiber.Map{
		"results": nil,
	})
}

// internalServerError - Just log and return fiber.ErrInternalServerError
func internalServerError(err error) error {
	log.Println(err)
	return fiber.ErrInternalServerError
}
