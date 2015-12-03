package etcd

import "net/url"

// prependScheme adds http:// to the front of URLs if no scheme is provided.
func prependSchemeToMachines(machines []string) []string {
	fixedmachines := make([]string, 0)

	for _, machine := range machines {
		u, err := url.Parse(machine)
		if err != nil {
			panic(err)
		}

		if u.Scheme == "" {
			u.Scheme = "http"
		}

		fixedmachines = append(fixedmachines, u.String())
	}

	return fixedmachines
}
