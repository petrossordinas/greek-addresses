package service

import (
	"encoding/json"
	"gr_addresses/database"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

// Maximum results per query
const limit = 40

// applyLimit caps a query at limit results, unless noLimit is set, in which
// case the caller takes responsibility for the (potentially large) result set.
func applyLimit(dbq *gorm.DB, noLimit bool) *gorm.DB {
	if noLimit {
		return dbq
	}
	return dbq.Limit(limit)
}

// likePattern builds a SQL LIKE pattern from a raw query string. If the
// caller already included a '%' wildcard anywhere in qry, it is used as-is
// (allowing prefix, suffix, or substring searches). Otherwise qry is treated
// as a prefix search, matching the historical behavior.
func likePattern(qry string) string {
	if strings.Contains(qry, "%") {
		return qry
	}
	return qry + "%"
}

// Prefecture -
type Prefecture struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// City -
type City struct {
	ID           uint        `json:"id"`
	PrefectureID uint        `json:"prefecture_id"`
	Prefecture   *Prefecture `json:"prefecture,omitempty"`
	Name         string      `json:"name"`
}

// Zipcode -
type Zipcode struct {
	ID           uint        `json:"id"`
	CityID       uint        `json:"city_id"`
	City         *City       `json:"city,omitempty"`
	PrefectureID uint        `json:"prefecture_id"`
	Prefecture   *Prefecture `json:"prefecture,omitempty"`
	Zipcode      string      `json:"zipcode"`
}

// Street -
type Street struct {
	ID           uint        `json:"id"`
	ZipcodeID    uint        `json:"zipcode_id"`
	Zipcode      *Zipcode    `json:"zipcode,omitempty"`
	CityID       uint        `json:"city_id"`
	City         *City       `json:"city,omitempty"`
	PrefectureID uint        `json:"prefecture_id"`
	Prefecture   *Prefecture `json:"prefecture,omitempty"`
	Name         string      `json:"name"`
	Ranges       string      `json:"ranges"`
}

// AddressSearch -
type AddressSearch struct {
	prefecture   Prefecture
	city         City
	zipcode      Zipcode
	streetNumber string
}

// FilterPrefecture -
func (a *AddressSearch) FilterPrefecture(qry string, noLimit bool) ([]Prefecture, error) {
	var prefectures []Prefecture
	result := applyLimit(database.DBConn.Where("name LIKE ?", likePattern(qry)), noLimit).
		Find(&prefectures)
	if result.Error != nil {
		return nil, result.Error
	}
	return prefectures, nil
}

// FilterCity -
func (a *AddressSearch) FilterCity(qry string, noLimit bool) ([]City, error) {
	var cities []City
	result := applyLimit(database.DBConn.Preload("Prefecture").
		Where("name LIKE ?", likePattern(qry)), noLimit).
		Find(&cities)
	if result.Error != nil {
		return nil, result.Error
	}
	return cities, nil
}

// FilterZipcode -
func (a *AddressSearch) FilterZipcode(qry string, noLimit bool) ([]Zipcode, error) {
	var zipcodes []Zipcode
	result := applyLimit(database.DBConn.Preload("Prefecture").
		Preload("City").
		Where("zipcode LIKE ?", likePattern(qry)), noLimit).
		Find(&zipcodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return zipcodes, nil
}

// GetAllZipcodes -
func (a *AddressSearch) GetAllZipcodes() ([]Zipcode, error) {
	var zipcodes []Zipcode
	result := database.DBConn.Preload("Prefecture").
		Preload("City").
		Order("zipcode").
		Find(&zipcodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return zipcodes, nil
}

// FilterStreet - qry is the street name (with optional street number).
// prefectureQry, cityQry and zipcodeQry are optional wildcard filters on the
// related prefecture/city/zipcode name, applied via joins in addition to
// (and combinable with) the ID-based filters already resolved on a.
func (a *AddressSearch) FilterStreet(qry, prefectureQry, cityQry, zipcodeQry string, noLimit bool) ([]Street, error) {
	var streets []Street
	dbq := database.DBConn
	if a.prefecture.ID != 0 {
		dbq = dbq.Where("streets.prefecture_id = ?", a.prefecture.ID)
	}
	if prefectureQry != "" {
		dbq = dbq.Joins("JOIN prefectures ON prefectures.id = streets.prefecture_id").
			Where("prefectures.name LIKE ?", likePattern(prefectureQry))
	}
	if a.city.ID != 0 {
		dbq = dbq.Where("streets.city_id = ?", a.city.ID)
	}
	if cityQry != "" {
		dbq = dbq.Joins("JOIN cities ON cities.id = streets.city_id").
			Where("cities.name LIKE ?", likePattern(cityQry))
	}
	if a.zipcode.ID != 0 {
		dbq = dbq.Where("streets.zipcode_id = ?", a.zipcode.ID)
	}
	if zipcodeQry != "" {
		dbq = dbq.Joins("JOIN zipcodes ON zipcodes.id = streets.zipcode_id").
			Where("zipcodes.zipcode LIKE ?", likePattern(zipcodeQry))
	}
	// Determine qry street name and street number with regex.
	// Street number is any digit that may be followed by a single character at the end of the string
	r := regexp.MustCompile(`\d+.?$`)
	streetNumberStr := r.FindString(qry)
	streetName := strings.Trim(r.ReplaceAllString(qry, ""), " ")
	result := applyLimit(dbq.Preload("Prefecture").
		Preload("City").
		Preload("Zipcode").
		Where("streets.name LIKE ?", likePattern(streetName)), noLimit).
		Find(&streets)
	if result.Error != nil {
		return nil, result.Error
	}
	if len(streets) > 0 && streetNumberStr != "" {
		return filterStreetsByNumber(streets, streetNumberStr), nil
	}
	return streets, nil
}

// filterStreetsByNumber narrows streets down to those whose Ranges include
// streetNumberStr, a street number that may end with a single letter (e.g.
// "12Α"). A street with no Ranges matches any number. Each matching street's
// Name has streetNumberStr appended.
func filterStreetsByNumber(streets []Street, streetNumberStr string) []Street {
	// We need to parse the street number to an int.
	streetNumber, _ := strconv.ParseInt(streetNumberStr, 10, 32)
	// If the street number parses to 0, means we have a letter at the end, so retry by
	// removing the last rune of the street number string (it may be a
	// multi-byte character, e.g. a Greek letter, so trim by rune, not byte)
	if streetNumber == 0 {
		_, size := utf8.DecodeLastRuneInString(streetNumberStr)
		streetNumber, _ = strconv.ParseInt(streetNumberStr[:len(streetNumberStr)-size], 10, 32)
	}
	// We will return the found streets in this array instead of streets
	var foundStreets []Street
	for _, street := range streets {
		street.Name = street.Name + " " + streetNumberStr
		// If the street does not have any number ranges, return it as all street numbers
		// match this street
		if street.Ranges == "" {
			foundStreets = append(foundStreets, street)
			continue
		}
		// The ranges are in JSON form, with one or two fields "odd" or "even", each having
		// and array of json objects { from: number } and { to: number}
		var r map[string][]map[string]string
		e := json.Unmarshal([]byte(street.Ranges), &r)
		if e != nil {
			log.Println(e)
		}
		// Based on the street number, we determine if we are looking at the odd numbered or even
		// numbered side
		side := "even"
		if streetNumber%2 == 1 {
			side = "odd"
		}
		// If we have a range of numbers for the determined side
		numberRange, ok := r[side]
		if ok {
			// Go through all the number ranges for that side
			for _, v := range numberRange {
				// If the to field is empty, set it to 999 as there are no larger street numbers
				if v["to"] == "" {
					v["to"] = "999"
				}
				// We need those numbers as ints
				from, _ := strconv.ParseInt(v["from"], 10, 32)
				to, _ := strconv.ParseInt(v["to"], 10, 32)
				// If the street number is within the from and to range, the street matches
				if streetNumber >= from && streetNumber <= to {
					foundStreets = append(foundStreets, street)
				}
			}
		}
	}
	return foundStreets
}

// FindPrefectureByID -
func (a *AddressSearch) FindPrefectureByID(id string) (*AddressSearch, error) {
	var prefecture Prefecture
	result := database.DBConn.Where("id = ?", id).First(&prefecture)
	if result.Error != nil {
		return nil, result.Error
	}
	a.prefecture = prefecture
	return a, nil
}

// FindCityByID -
func (a *AddressSearch) FindCityByID(id string) (*AddressSearch, error) {
	var city City
	result := database.DBConn.Where("id = ?", id).First(&city)
	if result.Error != nil {
		return nil, result.Error
	}
	a.city = city
	return a, nil
}

// FindZipcodeByID -
func (a *AddressSearch) FindZipcodeByID(id string) (*AddressSearch, error) {
	var zipcode Zipcode
	result := database.DBConn.Where("id = ?", id).First(&zipcode)
	if result.Error != nil {
		return nil, result.Error
	}
	a.zipcode = zipcode
	return a, nil
}
