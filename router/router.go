package router

import (
	"errors"
	"gr_addresses/service"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
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
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))
	app.Get("/search", searchHandler)
}

// Search -
func searchHandler(c *fiber.Ctx) error {
	service := new(service.AddressSearch)
	streetQry := c.Query("street")
	noLimit := c.Query("nolimit") != ""

	// When searching for a street, prefecture/city/zipcode become wildcard
	// filters on that search instead of standalone lookups.
	if streetQry == "" {
		// Filter by prefecture
		if c.Query("prefecture") != "" && c.Query("prefecture_id") == "" {
			prefectures, err := service.FilterPrefecture(c.Query("prefecture"), noLimit)
			if err != nil {
				return internalServerError(err)
			}
			return c.JSON(fiber.Map{
				"results": prefectures,
			})
		}

		// Filter by city
		if c.Query("city") != "" && c.Query("city_id") == "" {
			cities, err := service.FilterCity(c.Query("city"), noLimit)
			if err != nil {
				return internalServerError(err)
			}
			return c.JSON(fiber.Map{
				"results": cities,
			})
		}

		// Get all zip codes
		if c.Query("zipcode") == "all" {
			zipcodes, err := service.GetAllZipcodes()
			if err != nil {
				return internalServerError(err)
			}
			return c.JSON(fiber.Map{
				"results": zipcodes,
			})
		}

		// Filter by zip code
		if c.Query("zipcode") != "" && c.Query("zipcode_id") == "" {
			zipcodes, err := service.FilterZipcode(c.Query("zipcode"), noLimit)
			if err != nil {
				return internalServerError(err)
			}
			return c.JSON(fiber.Map{
				"results": zipcodes,
			})
		}
	}

	// Find prefecture by id
	if c.Query("prefecture_id") != "" {
		var err error
		service, err = service.FindPrefectureByID(c.Query("prefecture_id"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return emptyResults(c)
			}
			return internalServerError(err)
		}
	}

	// Find city by id
	if c.Query("city_id") != "" {
		var err error
		service, err = service.FindCityByID(c.Query("city_id"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return emptyResults(c)
			}
			return internalServerError(err)
		}
	}

	// Find zipcode by id
	if c.Query("zipcode_id") != "" {
		var err error
		service, err = service.FindZipcodeByID(c.Query("zipcode_id"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return emptyResults(c)
			}
			return internalServerError(err)
		}
	}

	// Filter by street
	if streetQry != "" {
		streets, err := service.FilterStreet(streetQry, c.Query("prefecture"), c.Query("city"), c.Query("zipcode"), noLimit)
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

// emptyResults - Return a 200 with an empty results array, e.g. when an id
// lookup found no matching record.
func emptyResults(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"results": []any{},
	})
}
