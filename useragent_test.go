package publiccode

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestBuildUserAgent(t *testing.T) {
	cases := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{
			name: "no build information",
			info: nil,
			want: "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "released CLI: this module is the main one",
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "v5.4.3"}},
			want: "publiccode-parser-go/5.4.3 (+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "unstamped build of the CLI",
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "(devel)"}},
			want: "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "embedded as a library: this module is a dependency",
			info: &debug.BuildInfo{
				Main: debug.Module{Path: "example.org/harvester", Version: "v1.0.0"},
				Deps: []*debug.Module{
					{Path: "example.org/other", Version: "v0.1.0"},
					{Path: modulePath, Version: "v5.4.3"},
				},
			},
			want: "publiccode-parser-go/5.4.3 (+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "pseudo-version identifies the commit and is kept",
			info: &debug.BuildInfo{
				Main: debug.Module{Path: "example.org/harvester", Version: "v1.0.0"},
				Deps: []*debug.Module{{Path: modulePath, Version: "v5.4.4-0.20260316100201-5dd490bc4896"}},
			},
			want: "publiccode-parser-go/5.4.4-0.20260316100201-5dd490bc4896 " +
				"(+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "CLI installed with \"go install\": the main module is the command",
			info: &debug.BuildInfo{Main: debug.Module{Path: modulePath + "/publiccode-parser", Version: "v5.4.3"}},
			want: "publiccode-parser-go/5.4.3 (+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "release build of the CLI: no module version in the build information",
			info: &debug.BuildInfo{Main: debug.Module{Path: "command-line-arguments"}},
			want: "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
		},
		{
			name: "this module is nowhere to be found",
			info: &debug.BuildInfo{Main: debug.Module{Path: "example.org/harvester", Version: "v1.0.0"}},
			want: "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildUserAgent(tc.info); got != tc.want {
				t.Errorf("buildUserAgent() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUserAgentForVersion(t *testing.T) {
	cases := map[string]string{
		"5.4.3":   "publiccode-parser-go/5.4.3 (+https://github.com/italia/publiccode-parser-go)",
		"v5.4.3":  "publiccode-parser-go/5.4.3 (+https://github.com/italia/publiccode-parser-go)",
		"devel":   "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
		"(devel)": "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
		"":        "publiccode-parser-go/devel (+https://github.com/italia/publiccode-parser-go)",
	}

	for version, want := range cases {
		if got := UserAgentForVersion(version); got != want {
			t.Errorf("UserAgentForVersion(%q) = %q, want %q", version, got, want)
		}
	}
}

func TestUserAgentNamesThisProject(t *testing.T) {
	got := UserAgent()

	if !strings.HasPrefix(got, productName+"/") {
		t.Errorf("UserAgent() = %q, want it to start with %q", got, productName+"/")
	}
	if !strings.Contains(got, projectURL) {
		t.Errorf("UserAgent() = %q, want it to point at %q", got, projectURL)
	}
}

// userAgentRecorder is a server recording the User-Agent of every request it
// receives.
type userAgentRecorder struct {
	*httptest.Server

	mu   sync.Mutex
	seen []string
}

// newUserAgentRecorder starts a recording server, stopped when the test ends.
func newUserAgentRecorder(t *testing.T) *userAgentRecorder {
	t.Helper()

	recorder := &userAgentRecorder{}
	recorder.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.mu.Lock()
		recorder.seen = append(recorder.seen, r.UserAgent())
		recorder.mu.Unlock()

		w.Header().Set("Content-Type", "text/yaml")
		_, _ = w.Write([]byte("publiccodeYmlVersion: \"0.4\"\n"))
	}))

	t.Cleanup(recorder.Close)

	return recorder
}

// userAgents returns the User-Agents recorded so far.
func (r *userAgentRecorder) userAgents() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.seen)
}

// TestUserAgentIsSentOnEveryRequestPath covers the two ways the parser reaches
// the network: the request it builds itself in Parse(), and the ones built by
// httpclient-lib-go during the external checks.
func TestUserAgentIsSentOnEveryRequestPath(t *testing.T) {
	cases := map[string]string{
		"default":  UserAgent(),
		"override": "harvester/2.0 (+https://example.org)",
	}

	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			srv := newUserAgentRecorder(t)

			config := ParserConfig{AllowNetworkToPrivateHosts: true}
			if name == "override" {
				config.UserAgent = want
			}

			p, err := NewParser(config)
			if err != nil {
				t.Fatal(err)
			}

			// Parse() builds its request itself.
			_, _ = p.Parse(srv.URL + "/publiccode.yml")

			// The external checks go through httpclient-lib-go.
			u, err := url.Parse(srv.URL + "/anything")
			if err != nil {
				t.Fatal(err)
			}

			if _, err = p.isReachable(*u); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			seen := srv.userAgents()
			if len(seen) < 2 {
				t.Fatalf("expected both request paths to reach the server, got %d request(s)", len(seen))
			}

			for _, got := range seen {
				if got != want {
					t.Errorf("User-Agent received by the server: got %q, want %q", got, want)
				}
			}
		})
	}
}
