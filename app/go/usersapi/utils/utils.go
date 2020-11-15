package utils

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
)

// GetDB with the given configuration details.
func GetDB(driver, username, password, port, host, database string) (*gorm.DB, error) {
	var err error
	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", host, port, username, database, password)
	DB, err := gorm.Open(driver, DBURL)
	if err != nil {
		fmt.Printf("Cannot connect to %s database", driver)
		log.Fatal("Error:", err)
	} else {
		fmt.Printf("Connected to the %s database \n", driver)
	}

	return DB, nil
}