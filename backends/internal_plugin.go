package backends

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
	"github.com/kelseyhightower/confd/builtin/databases/consul"
	"github.com/kelseyhightower/confd/builtin/databases/dynamodb"
	"github.com/kelseyhightower/confd/builtin/databases/env"
	"github.com/kelseyhightower/confd/builtin/databases/etcd"
	"github.com/kelseyhightower/confd/builtin/databases/rancher"
	"github.com/kelseyhightower/confd/builtin/databases/redis"
	"github.com/kelseyhightower/confd/builtin/databases/stackengine"
	"github.com/kelseyhightower/confd/builtin/databases/vault"
	"github.com/kelseyhightower/confd/builtin/databases/zookeeper"
	confdplugin "github.com/kelseyhightower/confd/plugin"
)

var InternalDatabases = map[string]confdplugin.DatabaseFunc{
	"env":         env.Database,
	"consul":      consul.Database,
	"dynamodb":    dynamodb.Database,
	"etcd":        etcd.Database,
	"rancher":     rancher.Database,
	"redis":       redis.Database,
	"stackengine": stackengine.Database,
	"zookeeper":   zookeeper.Database,
	"vault":       vault.Database,
}

const CONFDSPACE = "-CONFDSPACE-"

// BuildPluginCommandString builds a special string for executing internal
// plugins. It has the following format:
//
// 	/path/to/confd-CONFDSPACE-internal-plugin-CONFDSPACE-confd-database-env
//
// We split the string on -CONFDSPACE- to build the command executor. The reason we
// use -CONFDSPACE- is so we can support spaces in the /path/to/confd part.
func BuildPluginCommandString(pluginType, pluginName string) (string, error) {
	path, err := osext.Executable()
	if err != nil {
		return "", err
	}
	parts := []string{path, "internal-plugin", pluginType, pluginName}
	return strings.Join(parts, CONFDSPACE), nil
}

func RunPlugin(args []string) int {
	if len(args) != 2 {
		log.Printf("Wrong number of args; expected: confd internal-plugin pluginType pluginName")
		return 1
	}

	pluginType := args[0]
	pluginName := args[1]

	log.SetPrefix(fmt.Sprintf("%s-%s (internal) ", pluginName, pluginType))

	switch pluginType {
	case confdplugin.DatabasePluginName:
		pluginFunc, found := InternalDatabases[pluginName]
		if !found {
			log.Printf("[ERROR] Could not load database: %s", pluginName)
			return 1
		}
		log.Printf("[INFO] Starting database plugin %s", pluginName)
		confdplugin.Serve(&confdplugin.ServeOpts{
			DatabaseFunc: pluginFunc,
		})
	default:
		log.Printf("[ERROR] Invalid plugin type %s", pluginType)
		return 1
	}

	return 0
}

// Discover plugins located on disk, and fall back on plugins baked into the
// confd binary.
//
// We look in the following places for plugins:
//
// 1. Path where confd is installed
// 2. Path where confd is invoked
//
// Whichever file is discoverd LAST wins.
//
// Finally, we look at the list of plugins compiled into confd. If any of
// them has not been found on disk we use the internal version. This allows
// users to add / replace plugins without recompiling the main binary.
func Discover() (plugins map[string]string, err error) {
	// Look in the same directory as the confd executable, usually
	// /usr/local/bin. If found, this replaces what we found in the config path.
	exePath, err := osext.Executable()
	if err != nil {
		log.Printf("[ERROR] Error loading exe directory: %s", err)
	} else {
		if err = discover(filepath.Dir(exePath), &plugins); err != nil {
			return
		}
	}

	// Finally look in the cwd (where we are invoke confd). If found, this
	// replaces anything we found in the config / install paths.
	if err = discover(".", &plugins); err != nil {
		return
	}

	// Finally, if we have a plugin compiled into confd and we didn't find
	// a replacement on disk, we'll just use the internal version. Only do this
	// from the main process, or the log output will break the plugin handshake.
	for name, _ := range InternalDatabases {
		if path, found := plugins[name]; found {
			// Allow these warnings to be suppressed via CONFD_PLUGIN_DEV=1 or similar
			if os.Getenv("CONFD_PLUGIN_DEV") == "" {
				log.Printf("[WARN] %s overrides an internal plugin for %s-database.\n"+
					"  If you did not expect to see this message you will need to remove the old plugin.\n",
					path, name)
			}
		} else {
			cmd, err := BuildPluginCommandString(confdplugin.DatabasePluginName, name)
			if err != nil {
				return plugins, err
			}
			plugins[name] = cmd
		}
	}

	return plugins, nil
}

func discover(path string, m *map[string]string) error {
	var err error
	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	err = discoverSingle(
		filepath.Join(path, "confd-database-*"), m)
	if err != nil {
		return err
	}

	return nil
}

func discoverSingle(glob string, m *map[string]string) error {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	if *m == nil {
		*m = make(map[string]string)
	}

	for _, match := range matches {
		file := filepath.Base(match)

		// If the filename has a ".", trim up to there
		if idx := strings.Index(file, "."); idx >= 0 {
			file = file[:idx]
		}

		// Look for foo-bar-baz. The plugin name is "baz"
		parts := strings.SplitN(file, "-", 3)
		if len(parts) != 3 {
			continue
		}

		log.Printf("[DEBUG] Discovered plugin: %s = %s", parts[2], match)
		(*m)[parts[2]] = match
	}

	return nil
}

func pluginCmd(path string) *exec.Cmd {
	cmdPath := ""

	// If the path doesn't contain a separator, look in the same
	// directory as the confd executable first.
	if !strings.ContainsRune(path, os.PathSeparator) {
		exePath, err := osext.Executable()
		if err == nil {
			temp := filepath.Join(
				filepath.Dir(exePath),
				filepath.Base(path))

			if _, err := os.Stat(temp); err == nil {
				cmdPath = temp
			}
		}

		// If we still haven't found the executable, look for it
		// in the PATH.
		if v, err := exec.LookPath(path); err == nil {
			cmdPath = v
		}
	}

	// No plugin binary found, so try to use an internal plugin.
	if strings.Contains(path, CONFDSPACE) {
		parts := strings.Split(path, CONFDSPACE)
		return exec.Command(parts[0], parts[1:]...)
	}

	// If we still don't have a path, then just set it to the original
	// given path.
	if cmdPath == "" {
		cmdPath = path
	}

	// Build the command to execute the plugin
	return exec.Command(cmdPath)
}
