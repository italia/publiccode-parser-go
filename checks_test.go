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
	list := []testURL{
		{"https://github.com", Domain{Host: "github.com"}, true},
		{"https://githubs.com", Domain{Host: "github.com"}, false},
		{"https://github.org", Domain{Host: "github.com"}, false},
		{"https://github", Domain{Host: "github.com"}, false},
		{"http://github.com", Domain{Host: "github.com"}, true},
		{"http:/github.com", Domain{Host: "github.com"}, false},
		{"", Domain{Host: ""}, false},
		{"", Domain{}, false},
	}
	for _, l := range list {
		out := isHostInDomain(l.toTest, l.url)
		if out != l.res {
			t.Errorf("some evaluation went wrong %s with %s", l.toTest, l.url)
		}
	}
}
