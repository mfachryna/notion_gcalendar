package domain

import (
	"gorm.io/gorm"
)

type App struct {
	PostgreDB *gorm.DB
}
