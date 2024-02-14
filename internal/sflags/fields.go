package sflags

import (
	"fmt"
	"strings"
)

// Parse string into map splitting it into fields by coma delimiter and
// key & value by column
//
// Example:
// "destination:example.com,nexthop:example.com" becomes
// map[string]string{"destination":"example.com", "nexthop":"example.com"}
func parse(s string) (map[string]string, error) {
	fields := strings.Split(s, ",")

	m := make(map[string]string)
	for _, field := range fields {
		// TODO: What we do with "destination:https://example.com" ?
		ss := strings.Split(field, ":")
		if len(ss) != 2 {
			return nil, fmt.Errorf("failed to parse field: %s", field)
		}

		m[ss[0]] = ss[1]
	}

	return m, nil
}

func ParseSlice(ss []string) ([]map[string]string, error) {
	var sm []map[string]string

	for _, s := range ss {
		m, err := parse(s)
		if err != nil {
			return nil, err
		}
		sm = append(sm, m)
	}

	return sm, nil
}
