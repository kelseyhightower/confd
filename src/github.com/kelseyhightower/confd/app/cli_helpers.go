package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/kelseyhightower/confd/log"
)

type Flag struct {
	Field  reflect.StructField
	Name   string
	Usage  string
	Value  interface{}
	EnvVar string
}

func getValue(key string, m map[string]interface{}) (interface{}, error) {
	v, ok := m[key]
	if !ok {
		return nil, errors.New("unable to get '" + key + "' from map")
	}
	return v, nil
}

func getString(key string, m map[string]interface{}) (string, error) {
	v, err := getValue(key, m)
	if err != nil {
		return "", err
	}
	s, ok := v.(string)
	if !ok {
		return "", errors.New("unable to cast '" + key + "' to string")
	}
	return s, nil
}

func eachValue(v interface{}, f func(string)error) {
	if sv, ok := v.(string); ok {
		println(sv)
		for _, s := range strings.Split(sv, ",") {
			s = strings.TrimSpace(s)
			err := f(s)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
			}
		}
	}
}

func getCliFlag(f Flag) (cli.Flag, error) {
	var cflag cli.Flag = nil
	switch f.Field.Type.Kind() {
	case reflect.Bool:
		cflag = cli.BoolFlag{Name: f.Name, Usage: f.Usage, EnvVar: f.EnvVar}
		break
	case reflect.Int:
		i, _ := f.Value.(float64)
		cflag = cli.IntFlag{Name: f.Name, Usage: f.Usage, Value: int(i), EnvVar: f.EnvVar}
		break
	case reflect.String:
		s, _ := f.Value.(string)
		cflag = cli.StringFlag{Name: f.Name, Usage: f.Usage, Value: s, EnvVar: f.EnvVar}
		break
	case reflect.Slice:
		switch f.Field.Type.Elem().Kind() {
		case reflect.Int:
			is := &cli.IntSlice{}
			eachValue(f.Value, is.Set)
			cflag = cli.IntSliceFlag{Name: f.Name, Value: is, Usage: f.Usage, EnvVar: f.EnvVar}
			break
		case reflect.String:
			ss := &cli.StringSlice{}
		    eachValue(f.Value, ss.Set)
			cflag = cli.StringSliceFlag{Name: f.Name, Value: ss, Usage: f.Usage, EnvVar: f.EnvVar}
			break
		default:
			return nil, errors.New("unsupported slice element type for: " + f.Name)
		}
		break
	default:
		cflag = cli.GenericFlag{Name: f.Name, Usage: f.Usage, EnvVar: f.EnvVar}
		break
	}

	return cflag, nil
}

func getCliFlags(flags []Flag) []cli.Flag {
	cflags := make([]cli.Flag, 0)
	for _, f := range flags {
		cf, err := getCliFlag(f)
		if err != nil {
			log.Fatal(err.Error())
		}
		cflags = append(cflags, cf)
	}
	return cflags
}

func getFlagsFromType(t reflect.Type) []Flag {
	flags := make([]Flag, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		data := f.Tag.Get("cli")
		if data == "" {
			continue
		}

		var m map[string]interface{}
		err := json.Unmarshal([]byte(data), &m)
		if err != nil {
			log.Fatal(err.Error())
		}

		name, err := getString("name", m)
		if err != nil {
			log.Fatal(err.Error())
		}
		usage, _ := getString("name", m)
		value, _ := getValue("value", m)
		envvar, _ := getString("envvar", m)

		flags = append(flags, Flag{f, name, usage, value, envvar})
	}
	return flags
}

func ctxIsSet(c *cli.Context, isGlobal bool) func(string) bool {
	if isGlobal {
		return c.GlobalIsSet
	}
	return c.IsSet
}

func ctxBool(c *cli.Context, isGlobal bool) func(string) bool {
	if isGlobal {
		return c.GlobalBool
	}
	return c.Bool
}

func ctxInt(c *cli.Context, isGlobal bool) func(string) int {
	if isGlobal {
		return c.GlobalInt
	}
	return c.Int
}

func ctxString(c *cli.Context, isGlobal bool) func(string) string {
	if isGlobal {
		return c.GlobalString
	}
	return c.String
}

func ctxIntSlice(c *cli.Context, isGlobal bool) func(string) []int {
	if isGlobal {
		return c.GlobalIntSlice
	}
	return c.IntSlice
}

func ctxStringSlice(c *cli.Context, isGlobal bool) func(string) []string {
	if isGlobal {
		return c.GlobalStringSlice
	}
	return c.StringSlice
}

func overwriteWithCliFlags(flags []Flag, c *cli.Context, isGlobal bool, i interface{}) {
	for _, f := range flags {
		hasEnvVar := f.EnvVar != "" && os.Getenv(f.EnvVar) != ""
		if !ctxIsSet(c, isGlobal)(f.Name) && !hasEnvVar {
			continue
		}
		v := reflect.ValueOf(i).Elem().FieldByName(f.Field.Name)
		switch f.Field.Type.Kind() {
		case reflect.Bool:
			v.SetBool(ctxBool(c, isGlobal)(f.Name))
			break
		case reflect.Int:
			v.SetInt(int64(ctxInt(c, isGlobal)(f.Name)))
			break
		case reflect.String:
			v.SetString(ctxString(c, isGlobal)(f.Name))
			break
		case reflect.Slice:
			var slice interface{} = nil
			var count int = 0
			switch f.Field.Type.Elem().Kind() {
			case reflect.Int:
				is := ctxIntSlice(c, isGlobal)(f.Name)
				count = len(is)
				slice = interface{}(is)
				break
			case reflect.String:
				ss := ctxStringSlice(c, isGlobal)(f.Name)
				count = len(ss)
				slice = interface{}(ss)
				break
			}
			values := reflect.MakeSlice(v.Type(), count, count)
			reflect.Copy(values, reflect.ValueOf(slice))
			v.Set(values)
		default:
			break
		}
	}
}
