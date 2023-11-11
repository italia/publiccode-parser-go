package main

import (
	"testing"
)

type testURL struct {
	toTest string
	res    bool
}

func TestValidURLs(t *testing.T) {
	list := []testURL{
		{"../.", false},
		{"https://google.it", true},
		{"https//google.it", false},
		{"https://google.it/hello", true},
		{"htps://google.it", false},
		{"/tmp/a.yml", false},
		{"tmp/a.yml", false},
		{"a.yml", false},
	}
	for _, l := range list {
		out, _ := isValidURL(l.toTest)
		if out != l.res {
			t.Errorf("some evaluation went wrong %s", l.toTest)
		}
	}
}
