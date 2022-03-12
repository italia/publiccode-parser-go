package urlutil

import (
	"net/url"
)

// IsValidURL tests a string to determine if it is a well-structured url or not.
func IsValidURL(toTest string) (bool, *url.URL) {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false, nil
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false, nil
	}

	// Check it's an acceptable scheme
	switch u.Scheme {
	case "http":
	case "https":
	default:
		return false, nil
	}

	return true, u
}

func DisplayURL(u *url.URL) string {
	if u.Scheme == "file" {
		return u.Path
	}

	return u.String()
}
