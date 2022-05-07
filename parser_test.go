package publiccode

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
)

type testType struct {
	file string
	err  error
}

// Parse the YAML file passed as argument, using the current directory
// as base path.
//
// Return nil if the parsing succeded or an error if it failed.
func parse(file string) error {
	var p *Parser
	var err error
	if p, err = NewParserWithPath(file, "."); err != nil {
		return err
	}

	return p.Parse()
}

// Check all the YAML files matching the glob pattern and fail for each file
// with parsing or validation errors.
func checkValidFiles(pattern string, t *testing.T) {
	testFiles, _ := filepath.Glob(pattern)
	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			err := parse(file)
			if err != nil {
				t.Errorf("[%s] validation failed for valid file: %T - %s\n", file, err, err)
			}
		})
	}
}

func checkParseErrors(t *testing.T, err error, test testType) {
	if test.err == nil && err != nil {
		t.Errorf("[%s] unexpected error: %v\n", test.file, err)
	} else if test.err != nil && err == nil {
		t.Errorf("[%s] no error generated\n", test.file)
	} else if test.err != nil && err != nil {
		if !reflect.DeepEqual(test.err, err) {
			t.Errorf("[%s] wrong error generated:\n%T - %s\n- instead of:\n%T - %s", test.file, err, err, test.err, test.err)
		}
	}
}

func TestValidPreviousStandardVersion(t *testing.T) {
	file := "testdata/v0.2/valid/valid.minimal.yml"
	t.Run(file,  func(t *testing.T) {
		err := parse(file)
		checkParseErrors(t, err, testType{
			file, ValidationResults{
				ValidationWarning{"publiccodeYmlVersion", "'0.2' is not the latest version, use '0.3'", 1, 1},
			},
		})
	})
}
func TestInvalidTestcasesV0_3(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Can't get current working directory")
	}

	expected := map[string]error{
		// publiccodeYmlVersion
		"publiccodeYmlVersion_missing.yml": ValidationResults{ValidationError{"publiccodeYmlVersion", "required", 0, 0}},
		"publiccodeYmlVersion_invalid.yml": ValidationResults{ValidationError{
			"publiccodeYmlVersion", "must be one of the following: 0.2 0.2.0 0.2.1 0.2.2 0.3 0.3.0", 2, 1,
		}},
		"publiccodeYmlVersion_wrong_type.yml": ValidationResults{
			ValidationError{"publiccodeYmlVersion", "wrong type for this field", 2, 1},
			ValidationError{"publiccodeYmlVersion", "required", 2, 1}},

		// name
		"name_missing.yml": ValidationResults{ValidationError{"name", "required", 1, 1}},
		"name_nil.yml":     ValidationResults{ValidationError{"name", "required", 4, 1}},
		"name_wrong_type.yml": ValidationResults{
			ValidationError{"name", "wrong type for this field", 4, 1},
			ValidationError{"name", "required", 4, 1},
		},

		// applicationSuite
		"applicationSuite_wrong_type.yml": ValidationResults{ValidationError{"applicationSuite", "wrong type for this field", 4, 1}},

		// url
		"url_missing.yml": ValidationResults{ValidationError{"url", "required", 1, 1}},
		"url_wrong_type.yml": ValidationResults{
			ValidationError{"url", "wrong type for this field", 6, 1},
			ValidationError{"url", "'' not reachable: missing URL scheme", 6, 1},
			ValidationError{"url", "is not a valid code repository", 6, 1},
		},
		"url_invalid.yml": ValidationResults{
			ValidationError{"url", "'foobar' not reachable: missing URL scheme", 6, 1},
			ValidationError{"url", "is not a valid code repository", 6, 1},
		},

		// landingURL
		"landingURL_wrong_type.yml": ValidationResults{
			ValidationError{"landingURL", "wrong type for this field", 8, 1},
			ValidationError{"landingURL", "'' not reachable: missing URL scheme", 8, 1},
		},
		"landingURL_invalid.yml": ValidationResults{
			ValidationError{"landingURL", "'???' not reachable: missing URL scheme", 8, 1},
		},

		// isBasedOn
		"isBasedOn_wrong_type.yml": ValidationResults{
			ValidationError{"isBasedOn.foobar", "wrong type for this field", 10, 1},
		},
		"isBasedOn_bad_url_array.yml": ValidationResults{
			ValidationError{"isBasedOn", "wrong type for this field", 8, 1},
		},
		"isBasedOn_bad_url_string.yml": ValidationResults{
			ValidationError{"isBasedOn", "'???' not reachable: missing URL scheme", 8, 1},
		},

		// softwareVersion
		"softwareVersion_wrong_type.yml": ValidationResults{
			ValidationError{"softwareVersion", "wrong type for this field", 8, 1},
		},

		// releaseDate
		"releaseDate_missing.yml": ValidationResults{ValidationError{"releaseDate", "required", 1, 1}},
		"releaseDate_wrong_type.yml": ValidationResults{
			ValidationError{"releaseDate", "wrong type for this field", 8, 1},
			ValidationError{"releaseDate", "required", 8, 1},
		},
		"releaseDate_invalid.yml": ValidationResults{
			ValidationError{"releaseDate", "must be a date with format 'YYYY-MM-DD'", 8, 1},
		},

		// logo
		"logo_wrong_type.yml": ValidationResults{
			ValidationError{"logo", "wrong type for this field", 18, 1},
		},
		"logo_unsupported_extension.yml": ValidationResults{
			ValidationError{"logo", fmt.Sprintf("invalid file extension for: %s/logo.mpg", cwd), 18, 1},
		},
		"logo_missing_file.yml": ValidationResults{
			ValidationError{"logo", fmt.Sprintf("no such file: %s/no_such_file.png", cwd), 18, 1},
		},
		"logo_invalid_png.yml": ValidationResults{
			ValidationError{"logo", "image: unknown format", 18, 1},
		},

		// monochromeLogo
		"monochromeLogo_wrong_type.yml": ValidationResults{
			ValidationError{"monochromeLogo", "wrong type for this field", 18, 1},
		},
		"monochromeLogo_unsupported_extension.yml": ValidationResults{
			ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 18, 1},
			ValidationError{
				"monochromeLogo",
				fmt.Sprintf("invalid file extension for: %s/monochromeLogo.mpg", cwd),
				18,
				1,
			},
		},
		"monochromeLogo_missing_file.yml": ValidationResults{
			ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 18, 1},
			ValidationError{"monochromeLogo", fmt.Sprintf("no such file: %s/no_such_file.png", cwd), 18, 1},
		},
		"monochromeLogo_invalid_png.yml": ValidationResults{
			ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 18, 1},
			ValidationError{"monochromeLogo", "image: unknown format", 18, 1},
		},

		// inputTypes
		"inputTypes_invalid.yml": ValidationResults{
			ValidationError{"inputTypes[1]", "'foobar' is not a MIME type", 1, 1},
		},
		"inputTypes_wrong_type.yml": ValidationResults{
			ValidationError{"inputTypes.foobar", "wrong type for this field", 15, 1},
		},

		// outputTypes
		"outputTypes_invalid.yml": ValidationResults{
			ValidationError{"outputTypes[1]", "'foobar' is not a MIME type", 1, 1},
		},
		"outputTypes_wrong_type.yml": ValidationResults{
			ValidationError{"outputTypes.foobar", "wrong type for this field", 15, 1},
		},

		// platforms
		"platforms_missing.yml": ValidationResults{ValidationError{"platforms", "must be more than 0", 1, 1}},
		"platforms_wrong_type.yml": ValidationResults{
			ValidationError{"platforms", "wrong type for this field", 14, 1},
			ValidationError{"platforms", "must be more than 0", 14, 1},
		},

		// categories
		"categories_missing.yml": ValidationResults{
			ValidationError{"categories", "required", 1, 1},
		},
		"categories_nil.yml": ValidationResults{
			ValidationError{"categories", "required", 17, 1},
		},
		"categories_empty.yml": ValidationResults{
			ValidationError{"categories", "must be more than 0", 17, 1},
		},
		"categories_invalid.yml": ValidationResults{ValidationError{"categories[0]", "must be a valid category", 1, 1}},

		// usedBy
		"usedBy_wrong_type.yml": ValidationResults{
			ValidationError{"usedBy", "wrong type for this field", 14, 1},
		},

		// roadmap
		"roadmap_invalid.yml": ValidationResults{
			ValidationError{"roadmap", "'foobar' not reachable: missing URL scheme", 4, 1},
		},
		"roadmap_wrong_type.yml": ValidationResults{
			ValidationError{"roadmap", "wrong type for this field", 4, 1},
			ValidationError{"roadmap", "'' not reachable: missing URL scheme", 4, 1},
		},

		// developmentStatus
		"developmentStatus_missing.yml": ValidationResults{
			ValidationError{"developmentStatus", "required", 1, 1},
		},
		"developmentStatus_invalid.yml": ValidationResults{
			ValidationError{"developmentStatus", "must be one of the following: concept development beta stable obsolete", 21, 1},
		},
		"developmentStatus_wrong_type.yml": ValidationResults{
			ValidationError{"developmentStatus", "wrong type for this field", 21, 1},
			ValidationError{"developmentStatus", "required", 21, 1},
		},

		// softwareType
		"softwareType_missing.yml": ValidationResults{
			ValidationError{"softwareType", "required", 1, 1},
		},
		"softwareType_invalid.yml": ValidationResults{
			ValidationError{"softwareType", "must be one of the following: standalone/mobile standalone/iot standalone/desktop standalone/web standalone/backend standalone/other addon library configurationFiles", 22, 1},
		},
		"softwareType_wrong_type.yml": ValidationResults{
			ValidationError{"softwareType", "wrong type for this field", 22, 1},
			ValidationError{"softwareType", "required", 22, 1},
		},

		// intendedAudience
		// intendedAudience.*
		"intendedAudience_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience", "wrong type for this field", 18, 1},
		},
		"intendedAudience_countries_invalid_country.yml": ValidationResults{
			ValidationError{"intendedAudience.countries[1]", "must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code", 18, 5},
		},
		"intendedAudience_countries_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.countries", "wrong type for this field", 19, 1},
		},
		"intendedAudience_unsupportedCountries_invalid_country.yml": ValidationResults{
			ValidationError{"intendedAudience.unsupportedCountries[0]", "must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code", 18, 5},
		},
		"intendedAudience_unsupportedCountries_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.unsupportedCountries", "wrong type for this field", 19, 1},
		},
		"intendedAudience_scope_invalid_scope.yml": ValidationResults{
			ValidationError{"intendedAudience.scope[0]", "must be a valid scope", 18, 5},
		},
		"intendedAudience_scope_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.scope", "wrong type for this field", 19, 1},
		},

		// description
		// description.*
		"description_eng_features_missing.yml": ValidationResults{
			ValidationError{"description.eng.features", "must be more than 0", 27, 5},
		},
		"description_eng_features_empty.yml": ValidationResults{
			ValidationError{"description.eng.features", "must be more than 0", 45, 5},
		},
		"description_eng_localisedName_wrong_type.yml": ValidationResults{
			ValidationError{"description.eng.localisedName", "wrong type for this field", 26, 1},
		},
		"description_eng_genericName_missing.yml": ValidationResults{
			ValidationError{"description.eng.genericName", "required", 25, 5},
		},
		"description_eng_genericName_too_long.yml": ValidationResults{
			ValidationError{"description.eng.genericName", "must be less or equal than 35", 27, 5},
		},
		"description_eng_shortDescription_missing.yml": ValidationResults{
			ValidationError{"description.eng.shortDescription", "required", 25, 5},
		},
		"description_eng_shortDescription_too_short.yml": ValidationResults{
			ValidationError{"description.eng.shortDescription", "required", 25, 5},
		},
		"description_eng_longDescription_missing.yml": ValidationResults{
			ValidationError{"description.eng.longDescription", "required", 25, 5},
		},
		"description_eng_longDescription_too_long.yml": ValidationResults{
			ValidationError{"description.eng.longDescription", "must be less or equal than 10000", 33, 5},
		},
		"description_eng_longDescription_too_short.yml": ValidationResults{
			ValidationError{"description.eng.longDescription", "must be more or equal than 500", 33, 5},
		},
		"description_eng_longDescription_too_short_grapheme_clusters.yml": ValidationResults{
			ValidationError{"description.eng.longDescription", "must be more or equal than 500", 34, 5},
		},
		"description_eng_screenshots_missing_file.yml": ValidationResults{
			ValidationError{
				"description.eng.screenshots[0]",
				fmt.Sprintf("'no_such_file.png' is not an image: no such file : %s/no_such_file.png", cwd),
				25,
				5,
			},
		},
		"description_eng_awards_wrong_type.yml": ValidationResults{
			ValidationError{"description.eng.awards", "wrong type for this field", 46, 1},
		},
		"description_eng_videos_invalid.yml": ValidationResults{
			ValidationError{"description.eng.videos[0]", "'https://google.com' is not a valid video URL supporting oEmbed: invalid oembed link: https://google.com", 25, 5},
		},

		// legal
		// legal.*
		"legal_missing.yml": ValidationResults{ValidationError{"legal.license", "required", 0, 0}},
		"legal_wrong_type.yml": ValidationResults{
			ValidationError{"legal", "wrong type for this field", 47, 1},
			ValidationError{"legal.license", "required", 47, 8},
		},
		"legal_license_missing.yml": ValidationResults{ValidationError{"legal.license", "required", 47, 3}},
		"legal_license_invalid.yml": ValidationResults{ValidationError{
			"legal.license", "invalid license 'Invalid License'", 48, 3,
		}},
		"legal_authorsFile_missing_file.yml": ValidationResults{
			ValidationError{
				"legal.authorsFile",
				fmt.Sprintf("'%s/no_such_authors_file.txt' does not exist", cwd),
				43,
				3,
			},
		},

		// maintenance
		// maintenance.*
		"maintenance_type_missing.yml": ValidationResults{
			ValidationError{"maintenance.type", "required", 53, 3},
		},
		"maintenance_type_invalid.yml": ValidationResults{
			ValidationError{"maintenance.type", "must be one of the following: internal contract community none", 51, 3},
		},
		"maintenance_contacts_missing_with_type_community.yml": ValidationResults{
			ValidationError{"maintenance.contacts", "required_if Type community", 50, 3},
		},
		"maintenance_contacts_missing_with_type_internal.yml": ValidationResults{
			ValidationError{"maintenance.contacts", "required_if Type internal", 50, 3},
		},
		"maintenance_contacts_name_missing.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].name", "required", 0, 0},
		},
		"maintenance_contacts_email_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].email", "must be a valid email", 0, 0},
		},
		"maintenance_contractors_missing_with_type_contract.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "required_if Type contract", 50, 3},
		},
		"maintenance_contractors_invalid_type.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "wrong type for this field", 53, 1},
			ValidationError{"maintenance.contractors", "required_if Type contract", 53, 3},
		},
		"maintenance_contractors_name_missing.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].name", "required", 0, 0},
		},
		"maintenance_contractors_until_missing.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].until", "required", 0, 0},
		},
		"maintenance_contractors_until_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].until", "must be a date with format 'YYYY-MM-DD'", 0, 0},
		},
		"maintenance_contractors_email_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].email", "must be a valid email", 0, 0},
		},

		// localisation
		"localisation_availableLanguages_missing.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages", "required", 56, 3},
		},
		"localisation_availableLanguages_empty.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages", "must be more than 0", 58, 3},
		},
		"localisation_availableLanguages_invalid.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages[0]", "must be a valid BCP 47 language", 56, 3},
		},
		// TODO: Enable this test when https://github.com/italia/publiccode-parser-go/issues/47
		// is fixed
		//
		// "localisation_availableLanguages_invalid_bcp47.yml": ValidationResults{
		// 	ValidationError{"localisation.availableLanguages[0]", "must be a valid BCP 47 language", 56, 3},
		// },
		"localisation_localisationReady_missing.yml": ValidationResults{
			ValidationError{"localisation.localisationReady", "required", 58, 3},
		},

		// dependsOn
		"dependsOn_open_name_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.name", "wrong type for this field", 57, 1},
			ValidationError{"dependsOn.open[0].name", "required", 0, 0},
		},
		"dependsOn_open_versionMin_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.versionMin", "wrong type for this field", 58, 1},
		},
		"dependsOn_open_versionMax_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.versionMax", "wrong type for this field", 58, 1},
		},
		"dependsOn_open_version_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.version", "wrong type for this field", 58, 1},
		},
		"dependsOn_open_optional_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.optional", "wrong type for this field", 58, 1},
		},

		// it.*
		"it_riuso_codiceIPA_invalid.yml": ValidationResults{
			ValidationError{"it.riuso.codiceIPA", "must be a valid Italian Public Administration Code (iPA)", 56, 5},
		},

		// misc
		"file_encoding.yml": ValidationResults{ValidationError{"", "Invalid UTF-8", 0, 0}},
		"invalid_yaml.yml": ValidationResults{ValidationError{"", "yaml: did not find expected key", 18, 1}},
	}

	testFiles, _ := filepath.Glob("testdata/v0.3/invalid/*yml")
	for _, file := range testFiles {
		baseName := path.Base(file)
		if expected[baseName] == nil {
			t.Errorf("No expected data for file %s", baseName)
		}
		t.Run(file, func(t *testing.T) {
			err := parse(file)
			checkParseErrors(t, err, testType{file, expected[baseName]})
		})
	}
}

// Test v0.3 valid YAML testcases (testdata/v0.3/valid/).
func TestValidTestcasesV0_3(t *testing.T) {
	checkValidFiles("testdata/v0.3/valid/*.yml", t)
}

// Test publiccode.yml remote files for key errors.
func TestDecodeValueErrorsRemote(t *testing.T) {
	testRemoteFiles := []testType{
		{"https://raw.githubusercontent.com/italia/publiccode-editor/master/publiccode.yml", ValidationResults{
			ValidationWarning{"publiccodeYmlVersion", "'0.2' is not the latest version, use '0.3'", 1, 1},
		}},
	}

	for _, test := range testRemoteFiles {
		t.Run(fmt.Sprintf("%v", test.err), func(t *testing.T) {
			var p *Parser
			var err error

			if p, err = NewParser(test.file); err != nil {
				t.Errorf("Can't create parser for %s", test.file)
			}
			err = p.Parse()

			checkParseErrors(t, err, test)
		})
	}
}

func TestUrlMissingWithoutPath(t *testing.T) {
	expected := map[string]error{
		"url_missing.yml": ValidationResults{
			ValidationError{"url", "required", 1, 1},
		},
	}

	testFiles, _ := filepath.Glob("testdata/v0.3/invalid/url_missing.yml")
	for _, file := range testFiles {
		baseName := path.Base(file)
		if expected[baseName] == nil {
			t.Errorf("No expected data for file %s", baseName)
		}
		t.Run(file, func(t *testing.T) {
			parser, err := NewParser(file)
			if err != nil {
				t.Errorf("Can't create parser for %s", file)
			}
			err = parser.Parse()

			checkParseErrors(t, err, testType{file, expected[baseName]})
		})
	}
}

func TestIsReachable(t *testing.T) {
	var p Parser
	p.DisableNetwork = true

	u, _ := url.Parse("https://google.com/404")
	if reachable, _ := p.isReachable(*u); !reachable {
		t.Errorf("isReachable() returned false with DisableNetwork enabled")
	}
}

// Test that the exported YAML passes validation again, and that re-exporting it
// matches the first export (lossless roundtrip).
func TestExport(t *testing.T) {
	var p *Parser
	var err error
	if p, err = NewParser("testdata/v0.3/valid/valid.yml"); err != nil {
		t.Errorf("Can't create Parser: %v", err)
	}
	p.DisableNetwork = true

	err = p.Parse()
	if err != nil {
		t.Errorf("Failed to parse valid file: %v", err)
	}

	yaml1, err := p.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML: %v", err)
	}

	var p2 *Parser
	if p2, err = NewParser("/dev/null"); err != nil {
		t.Errorf("Can't create Parser: %v", err)
	}
	p2.DisableNetwork = true

	err = p2.ParseBytes(yaml1)
	if err != nil {
		t.Errorf("Failed to parse exported file: %v", err)
	}

	yaml2, err := p2.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML again: %v", err)
	}

	if !bytes.Equal(yaml1, yaml2) {
		t.Errorf("Exported YAML files do not match; roundtrip is not lossless")
	}
}

// Test the toURL function
func TestToURL(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Can't get current working directory")
	}

	expected := map[string]*url.URL{
		"file.txt":                              &url.URL{Scheme: "file", Path: fmt.Sprintf("%s/file.txt", cwd)},
		"/path/file.txt":                        &url.URL{Scheme: "file", Path: "/path/file.txt"},
		"https://developers.italia.it/":         &url.URL{Scheme: "https", Host: "developers.italia.it", Path: "/"},
		"https://developers.italia.it/file.txt": &url.URL{Scheme: "https", Host: "developers.italia.it", Path: "/file.txt"},
		"http://developers.italia.it/":          &url.URL{Scheme: "http", Host: "developers.italia.it", Path: "/"},
	}

	for in, out := range expected {
		var u *url.URL
		if u, err = toURL(in); err != nil {
			t.Errorf("%s: got error %v", in, err)
		}

		if *u != *out {
			t.Errorf("%s: expected %v got %v", in, out, u)
		}
	}
}
