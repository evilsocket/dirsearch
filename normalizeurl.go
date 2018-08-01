package dirsearch

import (
	"errors"
	"strings"
)

// NormalizeURL takes any domain or URL as input and normalizes
// it adding (if needed) default schema and path.
func NormalizeURL(base *string) (err error) {
	if *base == "" {
		return errors.New("URL is empty.")
	}

	// add schema
	if !strings.Contains(*base, "://") {
		*base = "http://" + *base
	}

	// add path
	if !strings.HasSuffix(*base, "/") {
		*base += "/"
	}

	return nil
}
