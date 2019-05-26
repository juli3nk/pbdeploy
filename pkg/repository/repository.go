package repository

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type Repository interface {
	GetURL(string, string) string
	Exists(string, string) (bool, error)
	Create(string, string, bool) error
}

type RepositoryFactory func(conf map[string]string) (Repository, error)

var repositoryFactories = make(map[string]RepositoryFactory)

func supportedDrivers() string {
	drivers := make([]string, 0, len(repositoryFactories))

	for d := range repositoryFactories {
		drivers = append(drivers, string(d))
	}

	sort.Strings(drivers)

	return strings.Join(drivers, ",")
}

func RegisterDriver(name string, factory RepositoryFactory) {
	if factory == nil {
		log.Panicf("Repository factory %s does not exist.", name)
	}

	if _, registered := repositoryFactories[name]; registered {
		log.Printf("Repository factory %s already registered. Ignoring.", name)
	}

	repositoryFactories[name] = factory
}

func NewDriver(driver string, config map[string]string) (Repository, error) {
	engineFactory, exists := repositoryFactories[driver]
	if exists {
		return engineFactory(config)
	}

	return nil, fmt.Errorf("The driver: %s is not supported. Supported drivers are %s", driver, supportedDrivers())
}
