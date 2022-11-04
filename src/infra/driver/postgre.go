package driver

import (
	"errors"
	"fmt"
	"notioncalendar/src/util/env_driver"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewPostgreConn(env env_driver.PostgreEnv) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		env.Host,
		env.Username,
		env.Password,
		env.Database,
		env.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   env.Schema + ".", // schema name
			SingularTable: false,
		}})
	if err != nil {
		return nil, err
	}

	if db.Raw("SELECT 1").RowsAffected < 0 {
		return nil, errors.New("Failed to run test query")
	} else {
		fmt.Println("DB Connected")
	}

	return db, nil
}
