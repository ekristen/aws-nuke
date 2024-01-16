package nuke

import (
	"fmt"
	"github.com/ekristen/libnuke/pkg/resource"
)

var cloudControlMapping = map[string]string{}

func GetCloudControlMapping() map[string]string {
	return cloudControlMapping
}

func MapCloudControl(typeName string) resource.RegisterOption {
	return func(name string, lister resource.Lister) {
		_, exists := cloudControlMapping[typeName]
		if exists {
			panic(fmt.Sprintf("a cloud control mapping for %s already exists", typeName))
		}

		cloudControlMapping[typeName] = name
	}
}
