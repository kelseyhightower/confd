package plugin

import (
	"github.com/kelseyhightower/confd/confd"
)

func testDatabaseFixed(p confd.Database) DatabaseFunc {
	return func() confd.Database {
		return p
	}
}
