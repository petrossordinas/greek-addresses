package service

import (
	"encoding/json"
	"gr_addresses/database"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Maximum results per query
const limit = 40

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
func (a *AddressSearch) FilterPrefecture(qry string) ([]Prefecture, error) {
	var prefectures []Prefecture
	result := database.DBConn.Where("name LIKE ?", qry+"%").
		Limit(limit).
		Find(&prefectures)
	if result.Error != nil {
		return nil, result.Error
	}
	return prefectures, nil
}

// FilterCity -
func (a *AddressSearch) FilterCity(qry string) ([]City, error) {
	var cities []City
	result := database.DBConn.Preload("Prefecture").
		Where("name LIKE ?", qry+"%").
		Limit(limit).
		Find(&cities)
	if result.Error != nil {
		return nil, result.Error
	}
	return cities, nil
}

// FilterZipcode -
func (a *AddressSearch) FilterZipcode(qry string) ([]Zipcode, error) {
	var zipcodes []Zipcode
	result := database.DBConn.Preload("Prefecture").
		Preload("City").
		Where("zipcode LIKE ?", qry+"%").
		Limit(limit).
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

// FilterStreet -
func (a *AddressSearch) FilterStreet(qry string) ([]Street, error) {
	var streets []Street
	dbq := database.DBConn
	if a.prefecture.ID != 0 {
		dbq = dbq.Where("prefecture_id = ?", a.prefecture.ID)
	}
	if a.city.ID != 0 {
		dbq = dbq.Where("city_id = ?", a.city.ID)
	}
	if a.zipcode.ID != 0 {
		dbq = dbq.Where("zipcode_id = ?", a.zipcode.ID)
	}
	// Determine qry street name and street number with regex.
	// Street number is any digit that may be followed by a single character at the end of the string
	r := regexp.MustCompile(`\d+.?$`)
	streetNumberStr := r.FindString(qry)
	streetName := strings.Trim(r.ReplaceAllString(qry, ""), " ")
	result := dbq.Preload("Prefecture").
		Preload("City").
		Preload("Zipcode").
		Where("name LIKE ?", streetName+"%").
		Limit(limit).
		Find(&streets)
	if result.Error != nil {
		return nil, result.Error
	}
	if len(streets) > 0 && streetNumberStr != "" {
		// We need to parse the street number to an int.
		streetNumber, _ := strconv.ParseInt(streetNumberStr, 10, 32)
		// If the street number parses to 0, means we have a letter at the end, so retry by
		// removing the last character of the street number string
		if streetNumber == 0 {
			streetNumber, _ = strconv.ParseInt(streetNumberStr[:len(streetNumberStr)-1], 10, 32)
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
		return foundStreets, nil
	}
	return streets, nil
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

func numberInRange(streetRanges string) bool {
	var jsonToMap map[string]interface{}
	err := json.Unmarshal([]byte(streetRanges), &jsonToMap)
	if err != nil {
		return false
	}
	return false
}
