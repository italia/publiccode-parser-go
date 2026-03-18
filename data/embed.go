package data

import _ "embed"

var (
	//go:embed oembed_schemes.json
	OembedSchemes []byte
	//go:embed it/ipa_codes.txt
	ItIpaCodes string
)
