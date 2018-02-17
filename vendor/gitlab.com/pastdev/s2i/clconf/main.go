package main

import (
	"os"

	"gitlab.com/pastdev/s2i/clconf/clconf"
)

func main() {
	clconf.NewApp().Run(os.Args)
}
