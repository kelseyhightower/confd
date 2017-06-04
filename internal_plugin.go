package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/kardianos/osext"
	"github.com/kelseyhightower/confd/plugin"
)

const CONFDSPACE = "-CONFDSPACE-"

// BuildPluginCommandString builds a special string for executing internal
// plugins. It has the following format:
//
// 	/path/to/terraform-CONFDSPACE-internal-plugin-CONFDSPACE-terraform-provider-aws
//
// We split the string on -CONFDSPACE- to build the command executor. The reason we
// use -CONFDSPACE- is so we can support spaces in the /path/to/terraform part.
func BuildPluginCommandString(pluginType, pluginName string) (string, error) {
	terraformPath, err := osext.Executable()
	if err != nil {
		return "", err
	}
	parts := []string{terraformPath, "internal-plugin", pluginType, pluginName}
	return strings.Join(parts, CONFDSPACE), nil
}

func Run(args []string) int {
	if len(args) != 2 {
		log.Printf("Wrong number of args; expected: confd internal-plugin pluginType pluginName")
		return 1
	}

	pluginType := args[0]
	pluginName := args[1]

	log.SetPrefix(fmt.Sprintf("%s-%s (internal) ", pluginName, pluginType))

	switch pluginType {
	case "database":
		pluginFunc, found := InternalDatabases[pluginName]
		if !found {
			log.Printf("[ERROR] Could not load provider: %s", pluginName)
			return 1
		}
		log.Printf("[INFO] Starting provider plugin %s", pluginName)
		plugin.Serve(&plugin.ServeOpts{
			DatabaseFunc: pluginFunc,
		})
	default:
		log.Printf("[ERROR] Invalid plugin type %s", pluginType)
		return 1
	}

	return 0
}
