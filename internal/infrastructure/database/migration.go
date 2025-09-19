package database

import (
	"cleanArch_with_postgres/internal/entity"
	"fmt"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	migrate(db, &entity.User{})
	migrate(db, &entity.Blog{})
}

func migrate(db *gorm.DB, model interface{}) {
	err := db.AutoMigrate(model)
	if err != nil {
		fmt.Println(err)
	}
}
