package initializer

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnecttoDB() {
	var err error
	dsn := os.Getenv("DB")

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("Error Detail:", err)
		panic("Failed to connect to DB")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		panic("Failed to get sql.DB")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("Berhasil terhubung ke Database!")
}
