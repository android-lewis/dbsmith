package db

import (
	"fmt"

	"github.com/android-lewis/dbsmith/internal/models"
)

type driverRegistration struct {
	aliases     []string
	constructor func() Driver
}

var supportedDrivers = []driverRegistration{
	{
		aliases:     []string{"postgres", "postgresql"},
		constructor: func() Driver { return NewPostgresDriver() },
	},
	{
		aliases:     []string{"mysql", "mariadb"},
		constructor: func() Driver { return NewMySQLDriver() },
	},
	{
		aliases:     []string{"sqlite"},
		constructor: func() Driver { return NewSQLiteDriver() },
	},
}

type DriverFactory struct {
	drivers map[string]func() Driver
}

func NewDriverFactory() *DriverFactory {
	f := &DriverFactory{
		drivers: make(map[string]func() Driver),
	}

	for _, driver := range supportedDrivers {
		for _, alias := range driver.aliases {
			_ = f.Register(alias, driver.constructor)
		}
	}

	return f
}

func (f *DriverFactory) Register(dbType string, constructor func() Driver) error {
	if dbType == "" {
		return fmt.Errorf("database type cannot be empty")
	}

	if constructor == nil {
		return fmt.Errorf("constructor cannot be nil")
	}

	f.drivers[dbType] = constructor
	return nil
}

func (f *DriverFactory) Create(conn *models.Connection) (Driver, error) {
	if conn == nil {
		return nil, ErrInvalidConnection
	}

	constructor, ok := f.drivers[string(conn.Type)]
	if !ok {
		return nil, ErrInvalidConnection
	}

	return constructor(), nil
}

func (f *DriverFactory) IsSupported(dbType string) bool {
	_, ok := f.drivers[dbType]
	return ok
}

func (f *DriverFactory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.drivers))
	for k := range f.drivers {
		types = append(types, k)
	}
	return types
}
