package data

import _ "embed"

var (
	//go:embed oembed_providers.json
	OembedProviders []byte
	//go:embed it/ipa_codes.txt
	ItIpaCodes string
)
