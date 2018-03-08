package clconf

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli"
)

const (
	// Name is the name of this application
	Name = "clconf"
	// Version is the version of this application
	Version = "0.0.1"
)

// Makes dump unit testable as test classes can override print
// https://stackoverflow.com/a/26804949/516433
var print = fmt.Print

func cgetv(c *cli.Context) error {
	if err := c.Set("decrypt", "/"); err != nil {
		return cli.NewExitError(err, 1)
	}
	return getv(c)
}

func cliError(err error, exitCode int) cli.ExitCoder {
	if err != nil {
		if casted, ok := err.(cli.ExitCoder); ok {
			return casted
		}
		return cli.NewExitError(err, exitCode)
	}
	return nil
}

func csetv(c *cli.Context) error {
	if err := c.Set("encrypt", "true"); err != nil {
		return cli.NewExitError(err, 1)
	}
	return setv(c)
}

func dump(c *cli.Context, value interface{}, err cli.ExitCoder) cli.ExitCoder {
	if err != nil {
		return err
	}
	print(value)
	return nil
}

func getDefault(c *cli.Context) (string, bool) {
	if defaultValue := c.String("default"); defaultValue != "" {
		return defaultValue, true
	}
	return "", false
}

func getPath(c *cli.Context) string {
	valuePath := c.Args().First()

	if prefix := c.GlobalString("prefix"); prefix != "" {
		return path.Join(prefix, valuePath)
	} else if prefix, ok := os.LookupEnv("CONFIG_PREFIX"); ok {
		return path.Join(prefix, valuePath)
	}

	if valuePath == "" {
		return "/"
	}
	return valuePath
}

func getv(c *cli.Context) error {
	return dump(marshal(getValue(c)))
}

func getTemplate(c *cli.Context) (*cli.Context, *Template, cli.ExitCoder) {
	var tmpl *Template
	var err error
	templateString := c.String("template-string")
	if templateString != "" {
		secretAgent, _ := newSecretAgentFromCli(c)
		tmpl, err = NewTemplate("cli", templateString,
			&TemplateConfig{
				SecretAgent: secretAgent,
			})
	}
	if err == nil && tmpl == nil {
		templateBase64 := c.String("template-base64")
		if templateBase64 != "" {
			secretAgent, _ := newSecretAgentFromCli(c)
			tmpl, err = NewTemplateFromBase64("cli", templateBase64,
				&TemplateConfig{
					SecretAgent: secretAgent,
				})
		}
	}
	if err == nil && tmpl == nil {
		templateFile := c.String("template")
		if templateFile != "" {
			secretAgent, _ := newSecretAgentFromCli(c)
			tmpl, err = NewTemplateFromFile("cli", templateFile,
				&TemplateConfig{
					SecretAgent: secretAgent,
				})
		}
	}
	return c, tmpl, cliError(err, 1)
}

func getValue(c *cli.Context) (*cli.Context, interface{}, cli.ExitCoder) {
	path := getPath(c)
	config, err := load(c)
	if err != nil {
		return c, nil, cliError(err, 1)
	}
	value, ok := GetValue(config, path)
	if !ok {
		value, ok = getDefault(c)
		if !ok {
			return c, nil, cli.NewExitError(fmt.Sprintf("[%v] does not exist", path), 1)
		}
	}
	if decryptPaths := c.StringSlice("decrypt"); len(decryptPaths) > 0 {
		secretAgent, err := newSecretAgentFromCli(c)
		if err != nil {
			return c, nil, err
		}
		if stringValue, ok := value.(string); ok {
			if len(decryptPaths) != 1 || !(decryptPaths[0] == "" || decryptPaths[0] == "/") {
				return c, nil, cli.NewExitError("string value with non-root decrypt path", 1)
			}
			decrypted, err := secretAgent.Decrypt(stringValue)
			if err != nil {
				return c, nil, cliError(err, 1)
			}
			value = decrypted
		} else {
			err = cliError(secretAgent.DecryptPaths(value, decryptPaths...), 1)
			if err != nil {
				return c, nil, err
			}
		}
	}
	return c, value, nil
}

func getvFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  "decrypt",
			Usage: "A `list` of paths whose values needs to be decrypted",
		},
		cli.StringFlag{
			Name:  "default",
			Usage: "The value to be returned if the specified path does not exist (otherwise results in an error).",
		},
		cli.StringFlag{
			Name:  "template",
			Usage: "A go template file that will be executed against the resulting data.",
		},
		cli.StringFlag{
			Name:  "template-base64",
			Usage: "A base64 encoded string containing a go template that will be executed against the resulting data.",
		},
		cli.StringFlag{
			Name:  "template-string",
			Usage: "A string containing a go template that will be executed against the resulting data.",
		},
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "prefix",
			Usage: "Prepended to all getv/setv paths (env: CONFIG_PREFIX)",
		},
		cli.StringFlag{
			Name:  "secret-keyring",
			Usage: "Path to a gpg secring file (env: SECRET_KEYRING)",
		},
		cli.StringFlag{
			Name:  "secret-keyring-base64",
			Usage: "Base64 encoded gpg secring (env: SECRET_KEYRING_BASE64)",
		},
		cli.StringSliceFlag{
			Name:  "yaml",
			Usage: "A `list` of yaml files containing config (env: YAML_FILES).  If specified, YAML_FILES will be split on ',' and appended to this option.",
		},
		cli.StringSliceFlag{
			Name:  "yaml-base64",
			Usage: "A `list` of base 64 encoded yaml strings containing config (env: YAML_VARS).  If specified, YAML_VARS will be split on ',' and each value will be used to load a base64 string from an environtment variable of that name.  The values will be appended to this option.",
		},
	}
}

func load(c *cli.Context) (map[interface{}]interface{}, cli.ExitCoder) {
	config, err := LoadConfFromEnvironment(
		c.GlobalStringSlice("yaml"),
		c.GlobalStringSlice("yaml-base64"))
	return config, cliError(err, 1)
}

func loadForSetv(c *cli.Context) (string, map[interface{}]interface{}, cli.ExitCoder) {
	path, config, err := LoadSettableConfFromEnvironment(c.GlobalStringSlice("yaml"))
	return path, config, cliError(err, 1)
}

func marshal(c *cli.Context, value interface{}, err cli.ExitCoder) (*cli.Context, string, cli.ExitCoder) {
	if err != nil {
		return c, "", err
	}
	if _, tmpl, err := getTemplate(c); err != nil {
		return c, "", err
	} else if tmpl != nil {
		marshaled, err := tmpl.Execute(value)
		return c, marshaled, cliError(err, 1)
	} else if stringValue, ok := value.(string); ok {
		return c, stringValue, nil
	} else if mapValue, ok := value.(map[interface{}]interface{}); ok {
		marshaled, err := MarshalYaml(mapValue)
		return c, string(marshaled), cliError(err, 1)
	} else if arrayValue, ok := value.([]interface{}); ok {
		marshaled, err := MarshalYaml(arrayValue)
		return c, string(marshaled), cliError(err, 1)
	}
	return c, fmt.Sprintf("%v", value), err
}

// NewApp returns a new cli application instance ready to be run.
func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.UsageText = "clconf [global options] command [command options] [args...]"

	app.Flags = globalFlags()

	app.Commands = []cli.Command{
		{
			Name:      "cgetv",
			Usage:     "Get a secret value.  Simply an alias to `getv --decrypt /`",
			ArgsUsage: "PATH",
			Action:    cgetv,
			Flags:     getvFlags(),
		},
		{
			Name:      "getv",
			Usage:     "Get a value",
			ArgsUsage: "PATH",
			Action:    getv,
			Flags:     getvFlags(),
		},
		{
			Name:      "csetv",
			Usage:     "Set PATH to the encrypted value of VALUE in the file indicated by the global option --yaml (must be single valued).  Simply an alias to `setv --encrypt`",
			ArgsUsage: "PATH VALUE",
			Action:    csetv,
			Flags:     setvFlags(),
		},
		{
			Name:      "setv",
			Usage:     "Set PATH to VALUE in the file indicated by the global option --yaml (must be single valued).",
			ArgsUsage: "PATH VALUE",
			Action:    setv,
			Flags:     setvFlags(),
		},
	}

	app.Action = getv

	return app
}

func newSecretAgentFromCli(c *cli.Context) (*SecretAgent, cli.ExitCoder) {
	var err error
	var secretAgent *SecretAgent

	if keyBase64 := c.GlobalString("secret-keyring-base64"); keyBase64 != "" {
		secretAgent, err = NewSecretAgentFromBase64(keyBase64)
	} else if keyFile := c.GlobalString("secret-keyring"); keyFile != "" {
		secretAgent, err = NewSecretAgentFromFile(keyFile)
	} else if keyBase64, ok := os.LookupEnv("SECRET_KEYRING_BASE64"); ok {
		secretAgent, err = NewSecretAgentFromBase64(keyBase64)
	} else if keyFile, ok := os.LookupEnv("SECRET_KEYRING"); ok {
		secretAgent, err = NewSecretAgentFromFile(keyFile)
	} else {
		err = errors.New("requires --secret-keyring-base64, --secret-keyring, or SECRET_KEYRING")
	}

	return secretAgent, cliError(err, 1)
}

func setv(c *cli.Context) error {
	if c.NArg() != 2 {
		return cli.NewExitError("setv requires key and value args", 1)
	}

	path := getPath(c)
	value := c.Args().Get(1)
	file, config, err := loadForSetv(c)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to load config: %v", err), 1)
	}

	if c.Bool("encrypt") {
		secretAgent, err := newSecretAgentFromCli(c)
		if err != nil {
			return err
		}
		encrypted, encryptErr := secretAgent.Encrypt(value)
		if encryptErr != nil {
			return cli.NewExitError(
				fmt.Sprintf("Failed to encrypt value: %v", err), 1)
		}
		value = encrypted
	}

	if err := SetValue(config, path, value); err != nil {
		return cli.NewExitError(
			fmt.Sprintf("Failed to load config: %v", err), 1)
	}

	return cliError(SaveConf(config, file), 1)
}

func setvFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "encrypt",
			Usage: "Encrypt the value",
		},
	}
}
