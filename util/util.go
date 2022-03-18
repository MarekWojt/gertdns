package util

import "strings"

func ParseDomain(domain string) string {
	if !strings.HasSuffix(domain, ".") {
		return domain + "."
	}

	return domain
}
