package publiccode

import (
	"testing"
)

type testURL struct {
	url    string
	toTest Domain
	res    bool
}

func TestHostsInDomain(t *testing.T) {
	domain := Domain{
		UseTokenFor: []string{
			"github.com",
			"api.github.com",
			"raw.githubusercontent.com",
		},
	}

	list := []testURL{
		{"https://github.com", domain, true},
		{"https://githubs.com", domain, false},
		{"https://github.org", domain, false},
		{"https://github", domain, false},
		{"http://github.com", domain, true},
		{"http:/github.com", domain, false},
		{"https://api.github.com", domain, true},
		{"https://api.github.com/org/repo/file", domain, true},
		{"https://raw.githubusercontent.com/org/repo/master/pc.yml", domain, true},
		{"https://raw.githubusercontent.com", domain, true},
		{"", Domain{UseTokenFor: []string{}}, false},
		{"", Domain{}, false},
	}
	for _, l := range list {
		out := isHostInDomain(l.toTest, l.url)
		if out != l.res {
			t.Errorf("some evaluation went wrong %s with %s", l.toTest, l.url)
		}
	}
}
