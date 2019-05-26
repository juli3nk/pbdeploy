package file

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type Filer interface {
	String() string
	DeleteFiles() error
	CopyFiles() error
	CreateOrUpdateFiles() error
}

type FileFactory func(conf map[string][]string) (Filer, error)

var fileFactories = make(map[string]FileFactory)

func supportedDrivers() string {
	drivers := make([]string, 0, len(fileFactories))

	for d := range fileFactories {
		drivers = append(drivers, string(d))
	}

	sort.Strings(drivers)

	return strings.Join(drivers, ",")
}

func RegisterDriver(name string, factory FileFactory) {
	if factory == nil {
		log.Panicf("File factory %s does not exist.", name)
	}

	if _, registered := fileFactories[name]; registered {
		log.Printf("File factory %s already registered. Ignoring.", name)
	}

	fileFactories[name] = factory
}

func NewDriver(driver string, config map[string][]string) (Filer, error) {
	engineFactory, exists := fileFactories[driver]
	if exists {
		return engineFactory(config)
	}

	return nil, fmt.Errorf("The driver: %s is not supported. Supported drivers are %s", driver, supportedDrivers())
}
