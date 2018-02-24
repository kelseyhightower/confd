// Package template serves to isolate the large number of public functions
// created by template_funcs.go.  As this file is a copy of the version
// from confd, it is as unmodified as possible so as to allow for effective
// diffing when updates occur upstream.  This package also exposes 2 of
// the private functions so they can be consumed by the clconf package.
package template
