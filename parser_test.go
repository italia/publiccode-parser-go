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

	if p, err = NewDefaultParser(); err != nil {
		return err
	}

	_, err = p.Parse(file)

	return err
}

func parseNoNetwork(file string) error {
	var p *Parser
	var err error

	if p, err = NewParser(ParserConfig{DisableNetwork: true}); err != nil {
		return err
	}

	_, err = p.Parse(file)

	return err
}

// Check all the YAML files matching the glob pattern and fail for each file
// with parsing or validation errors.
func checkValidFiles(pattern string, t *testing.T) {
	testFiles, _ := filepath.Glob(pattern)
	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			if err := parse(file); err != nil {
				t.Errorf("[%s] validation failed for valid file: %T - %s\n", file, err, err)
			}
		})
	}
}

// Check all the YAML files matching the glob pattern and fail for each file
// with parsing or validation errors, with the network disabled.
func checkValidFilesNoNetwork(pattern string, t *testing.T) {
	testFiles, _ := filepath.Glob(pattern)
	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			if err := parseNoNetwork(file); err != nil {
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

func TestValidTestcasesV0_NoNetwork(t *testing.T) {
	checkValidFilesNoNetwork("testdata/v0/valid/no-network/*.yml", t)
}

func TestValidWithWarningTestcasesV0_NoNetwork(t *testing.T) {
	expected := map[string]error{
		"authorsFile.yml": ValidationResults{
			ValidationWarning{"legal.authorsFile", "This key is DEPRECATED and will be removed in the future", 72, 3},
		},
	}

	testFiles, _ := filepath.Glob("testdata/v0/valid_with_warnings/no-network/*yml")
	for _, file := range testFiles {
		baseName := path.Base(file)
		if expected[baseName] == nil {
			t.Errorf("No expected data for file %s", baseName)
		}
		t.Run(file, func(t *testing.T) {
			err := parseNoNetwork(file)
			checkParseErrors(t, err, testType{file, expected[baseName]})
		})
	}
}

func TestInvalidTestcasesV0_NoNetwork(t *testing.T) {
	expected := map[string]error{
		// logo
		"logo_missing_file.yml": ValidationResults{
			ValidationError{"logo", "no such file: no_such_file.png", 18, 1},
		},
		"logo_invalid_png.yml": ValidationResults{
			ValidationError{"logo", "image: unknown format", 18, 1},
		},

		// landingURL
		"landingURL_invalid.yml": ValidationResults{
			// Just a syntax check here, no check for reachability as network is disabled
			ValidationError{"landingURL", "landingURL must be an HTTP URL", 8, 1},
		},

		// monochromeLogo
		"monochromeLogo_invalid_png.yml": ValidationResults{
			ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 18, 1},
			ValidationError{"monochromeLogo", "image: unknown format", 18, 1},
		},
	}

	testFiles, _ := filepath.Glob("testdata/v0/invalid/no-network/*yml")
	for _, file := range testFiles {
		baseName := path.Base(file)
		if expected[baseName] == nil {
			t.Errorf("No expected data for file %s", baseName)
		}
		t.Run(file, func(t *testing.T) {
			err := parseNoNetwork(file)
			checkParseErrors(t, err, testType{file, expected[baseName]})
		})
	}
}

func TestInvalidTestcasesV0(t *testing.T) {
	expected := map[string]error{
		// publiccodeYmlVersion
		"publiccodeYmlVersion_missing.yml": ValidationResults{ValidationError{"publiccodeYmlVersion", "publiccodeYmlVersion is a required field", 0, 0}},
		"publiccodeYmlVersion_invalid.yml": ValidationResults{
			ValidationError{
				"publiccodeYmlVersion",
				"unsupported version: '1'. Supported versions: 0.2, 0.2.0, 0.2.1, 0.2.2, 0.3, 0.3.0, 0.4, 0.4.0",
				0,
				0,
			},
		},
		"publiccodeYmlVersion_wrong_type.yml": ValidationResults{
			ValidationError{"publiccodeYmlVersion", "wrong type for this field", 2, 1},
		},

		// name
		"name_missing.yml": ValidationResults{ValidationError{"name", "name is a required field", 1, 1}},
		"name_nil.yml":     ValidationResults{ValidationError{"name", "name is a required field", 4, 1}},
		"name_wrong_type.yml": ValidationResults{
			ValidationError{"name", "wrong type for this field", 4, 1},
			ValidationError{"name", "name is a required field", 4, 1},
		},

		// applicationSuite
		"applicationSuite_wrong_type.yml": ValidationResults{ValidationError{"applicationSuite", "wrong type for this field", 4, 1}},

		// url
		"url_missing.yml": ValidationResults{ValidationError{"url", "url is a required field", 1, 1}},
		"url_wrong_type.yml": ValidationResults{
			ValidationError{"url", "wrong type for this field", 6, 1},
			ValidationError{"url", "url must be a valid URL", 6, 1},
			ValidationError{"url", "'' not reachable: missing URL scheme", 6, 1},
			ValidationError{"url", "is not a valid code repository", 6, 1},
		},
		"url_invalid.yml": ValidationResults{
			ValidationError{"url", "url must be a valid URL", 6, 1},
			ValidationError{"url", "'foobar' not reachable: missing URL scheme", 6, 1},
			ValidationError{"url", "is not a valid code repository", 6, 1},
		},

		// landingURL
		"landingURL_wrong_type.yml": ValidationResults{
			ValidationError{"landingURL", "wrong type for this field", 8, 1},
			ValidationError{"landingURL", "landingURL must be an HTTP URL", 8, 1},
			ValidationError{"landingURL", "'' not reachable: missing URL scheme", 8, 1},
		},
		"landingURL_invalid.yml": ValidationResults{
			ValidationError{"landingURL", "landingURL must be an HTTP URL", 8, 1},
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
		"releaseDate_empty.yml": ValidationResults{ValidationError{"releaseDate", "releaseDate must be a date with format 'YYYY-MM-DD'", 8, 1}},
		"releaseDate_wrong_type.yml": ValidationResults{
			ValidationError{"releaseDate", "wrong type for this field", 8, 1},
			// FIXME: This isn't ideal, but it's a bug of the yaml library that deserializes
			// the field as a pointer to "" (two double quotes), instead of leaving it as nil.
			// It's still technically correct validation-wise.
			ValidationError{"releaseDate", "releaseDate must be a date with format 'YYYY-MM-DD'", 8, 1},
		},
		"releaseDate_invalid.yml": ValidationResults{
			ValidationError{"releaseDate", "releaseDate must be a date with format 'YYYY-MM-DD'", 8, 1},
		},

		// logo
		"logo_wrong_type.yml": ValidationResults{
			ValidationError{"logo", "wrong type for this field", 18, 1},
		},
		"logo_unsupported_extension.yml": ValidationResults{
			ValidationError{"logo", "invalid file extension for: https://raw.githubusercontent.com/italia/developers.italia.it/main/logo.mpg", 18, 1},
		},
		"logo_missing_file.yml": ValidationResults{
			ValidationError{"logo", "no such file: https://raw.githubusercontent.com/italia/developers.italia.it/main/no_such_file.png", 18, 1},
		},

		// monochromeLogo
		"monochromeLogo_wrong_type.yml": ValidationResults{
			ValidationError{"monochromeLogo", "wrong type for this field", 18, 1},
		},
		"monochromeLogo_unsupported_extension.yml": ValidationResults{
			ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 18, 1},
			ValidationError{
				"monochromeLogo",
				"invalid file extension for: https://raw.githubusercontent.com/italia/developers.italia.it/main/monochromeLogo.mpg",
				18,
				1,
			},
		},
		"monochromeLogo_missing_file.yml": ValidationResults{
			ValidationWarning{"monochromeLogo", "This key is DEPRECATED and will be removed in the future", 18, 1},
			ValidationError{"monochromeLogo", "no such file: https://raw.githubusercontent.com/italia/developers.italia.it/main/no_such_file.png", 18, 1},
		},

		// inputTypes
		"inputTypes_invalid.yml": ValidationResults{
			ValidationError{"inputTypes[1]", "inputTypes[1] is not a valid MIME type", 1, 1},
			ValidationWarning{"inputTypes", "This key is DEPRECATED and will be removed in the future", 14, 1},
		},
		"inputTypes_wrong_type.yml": ValidationResults{
			ValidationError{"inputTypes.foobar", "wrong type for this field", 15, 1},
		},

		// outputTypes
		"outputTypes_invalid.yml": ValidationResults{
			ValidationError{"outputTypes[1]", "outputTypes[1] is not a valid MIME type", 1, 1},
			ValidationWarning{"outputTypes", "This key is DEPRECATED and will be removed in the future", 14, 1},
		},
		"outputTypes_wrong_type.yml": ValidationResults{
			ValidationError{"outputTypes.foobar", "wrong type for this field", 15, 1},
		},

		// platforms
		"platforms_missing.yml": ValidationResults{ValidationError{"platforms", "platforms must contain more than 0 items", 1, 1}},
		"platforms_wrong_type.yml": ValidationResults{
			ValidationError{"platforms", "wrong type for this field", 9, 1},
			ValidationError{"platforms", "platforms must contain more than 0 items", 9, 1},
		},

		// categories
		"categories_missing.yml": ValidationResults{
			ValidationError{"categories", "categories is a required field", 1, 1},
		},
		"categories_nil.yml": ValidationResults{
			ValidationError{"categories", "categories is a required field", 12, 1},
		},
		"categories_empty.yml": ValidationResults{
			ValidationError{"categories", "categories must contain more than 0 items", 12, 1},
		},
		"categories_invalid.yml": ValidationResults{ValidationError{"categories[0]", "categories[0] must be a valid category (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/categories-list.rst)", 1, 1}},

		// usedBy
		"usedBy_wrong_type.yml": ValidationResults{
			ValidationError{"usedBy", "wrong type for this field", 14, 1},
		},

		// roadmap
		"roadmap_invalid.yml": ValidationResults{
			ValidationError{"roadmap", "roadmap must be an HTTP URL", 4, 1},
			ValidationError{"roadmap", "'foobar' not reachable: missing URL scheme", 4, 1},
		},
		"roadmap_wrong_type.yml": ValidationResults{
			ValidationError{"roadmap", "wrong type for this field", 4, 1},
			ValidationError{"roadmap", "roadmap must be an HTTP URL", 4, 1},
			ValidationError{"roadmap", "'' not reachable: missing URL scheme", 4, 1},
		},

		// developmentStatus
		"developmentStatus_missing.yml": ValidationResults{
			ValidationError{"developmentStatus", "developmentStatus is a required field", 1, 1},
		},
		"developmentStatus_invalid.yml": ValidationResults{
			ValidationError{"developmentStatus", "developmentStatus must be one of the following: \"concept\", \"development\", \"beta\", \"stable\" or \"obsolete\"", 16, 1},
		},
		"developmentStatus_wrong_type.yml": ValidationResults{
			ValidationError{"developmentStatus", "wrong type for this field", 16, 1},
			ValidationError{"developmentStatus", "developmentStatus is a required field", 16, 1},
		},

		// softwareType
		"softwareType_missing.yml": ValidationResults{
			ValidationError{"softwareType", "softwareType is a required field", 1, 1},
		},
		"softwareType_invalid.yml": ValidationResults{
			ValidationError{"softwareType", "softwareType must be one of the following: \"standalone/mobile\", \"standalone/iot\", \"standalone/desktop\", \"standalone/web\", \"standalone/backend\", \"standalone/other\", \"addon\", \"library\" or \"configurationFiles\"", 17, 1},
		},
		"softwareType_wrong_type.yml": ValidationResults{
			ValidationError{"softwareType", "wrong type for this field", 17, 1},
			ValidationError{"softwareType", "softwareType is a required field", 17, 1},
		},

		// intendedAudience
		// intendedAudience.*
		"intendedAudience_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience", "wrong type for this field", 18, 1},
		},
		"intendedAudience_countries_invalid_country.yml": ValidationResults{
			ValidationError{"intendedAudience.countries[1]", "countries[1] must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code", 18, 5},
			ValidationError{"intendedAudience.countries[2]", "countries[2] must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code", 18, 5},
		},
		"intendedAudience_countries_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.countries", "wrong type for this field", 19, 1},
		},
		"intendedAudience_unsupportedCountries_invalid_country.yml": ValidationResults{
			ValidationError{"intendedAudience.unsupportedCountries[0]", "unsupportedCountries[0] must be a valid lowercase ISO 3166-1 alpha-2 two-letter country code", 18, 5},
		},
		"intendedAudience_unsupportedCountries_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.unsupportedCountries", "wrong type for this field", 19, 1},
		},
		"intendedAudience_scope_invalid_scope.yml": ValidationResults{
			ValidationError{"intendedAudience.scope[0]", "scope[0] must be a valid scope (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/scope-list.rst)", 18, 5},
		},
		"intendedAudience_scope_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.scope", "wrong type for this field", 19, 1},
		},

		// description
		// description.*
		"description_invalid_language.yml": ValidationResults{
			ValidationError{"description", "description must be a valid BCP 47 language", 18, 1},
		},
		"description_en_features_missing.yml": ValidationResults{
			ValidationError{"description.en.features", "features must contain more than 0 items", 22, 5},
		},
		"description_en_features_empty.yml": ValidationResults{
			ValidationError{"description.en.features", "features must contain more than 0 items", 39, 5},
		},
		"description_en_localisedName_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.localisedName", "wrong type for this field", 21, 1},
		},
		"description_en_genericName_too_long.yml": ValidationResults{
			ValidationError{"description.en.genericName", "genericName must be a maximum of 35 characters in length", 22, 5},
			ValidationWarning{"description.en.genericName", "This key is DEPRECATED and will be removed in the future", 22, 5},
		},
		"description_en_shortDescription_missing.yml": ValidationResults{
			ValidationError{"description.en.shortDescription", "shortDescription is a required field", 20, 5},
		},
		"description_en_shortDescription_too_short.yml": ValidationResults{
			ValidationError{"description.en.shortDescription", "shortDescription is a required field", 20, 5},
		},
		"description_en_longDescription_missing.yml": ValidationResults{
			ValidationError{"description.en.longDescription", "longDescription is a required field", 20, 5},
		},
		"description_en_longDescription_too_long.yml": ValidationResults{
			ValidationError{"description.en.longDescription", "longDescription must be a maximum of 10000 characters in length", 27, 5},
		},
		"description_en_longDescription_too_short.yml": ValidationResults{
			ValidationError{"description.en.longDescription", "longDescription must be at least 150 characters in length", 27, 5},
		},
		"description_en_longDescription_too_short_grapheme_clusters.yml": ValidationResults{
			ValidationError{"description.en.longDescription", "longDescription must be at least 150 characters in length", 28, 5},
		},
		"description_en_documentation_invalid.yml": ValidationResults{
			ValidationError{"description.en.documentation", "documentation must be an HTTP URL", 25, 5},
			ValidationError{"description.en.documentation", "'not_a_url' not reachable: missing URL scheme", 25, 5},
		},
		"description_en_documentation_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.documentation", "wrong type for this field", 25, 1},
			ValidationError{"description.en.documentation", "documentation must be an HTTP URL", 25, 5},
			ValidationError{"description.en.documentation", "'' not reachable: missing URL scheme", 25, 5},
		},
		"description_en_apiDocumentation_invalid.yml": ValidationResults{
			ValidationError{"description.en.apiDocumentation", "apiDocumentation must be an HTTP URL", 41, 5},
			ValidationError{"description.en.apiDocumentation", "'abc' not reachable: missing URL scheme", 41, 5},
		},
		"description_en_apiDocumentation_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.apiDocumentation", "wrong type for this field", 43, 1},
			ValidationError{"description.en.apiDocumentation", "apiDocumentation must be an HTTP URL", 43, 5},
			ValidationError{"description.en.apiDocumentation", "'' not reachable: missing URL scheme", 43, 5},
		},
		"description_en_screenshots_missing_file.yml": ValidationResults{
			ValidationError{
				"description.en.screenshots[0]",
				"'no_such_file.png' is not an image: no such file : https://raw.githubusercontent.com/italia/developers.italia.it/main/no_such_file.png",
				20,
				5,
			},
		},
		"description_en_awards_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.awards", "wrong type for this field", 40, 1},
		},
		"description_en_videos_invalid.yml": ValidationResults{
			ValidationError{"description.en.videos[0]", "videos[0] must be an HTTP URL", 20, 5},
			ValidationError{"description.en.videos[0]", "'ABC' is not a valid video URL supporting oEmbed: invalid oEmbed link: ABC", 20, 5},
		},
		"description_en_videos_invalid_oembed.yml": ValidationResults{
			ValidationError{"description.en.videos[0]", "'https://google.com' is not a valid video URL supporting oEmbed: invalid oEmbed link: https://google.com", 20, 5},
		},

		// legal
		// legal.*
		"legal_missing.yml": ValidationResults{ValidationError{"legal.license", "license is a required field", 0, 0}},
		"legal_wrong_type.yml": ValidationResults{
			ValidationError{"legal", "wrong type for this field", 46, 1},
			ValidationError{"legal.license", "license is a required field", 46, 8},
		},
		"legal_license_missing.yml": ValidationResults{ValidationError{"legal.license", "license is a required field", 41, 3}},
		"legal_license_invalid.yml": ValidationResults{ValidationError{
			"legal.license", "license must be a valid license (see https://spdx.org/licenses)", 42, 3,
		}},
		"legal_authorsFile_missing_file.yml": ValidationResults{
			ValidationWarning{
				"legal.authorsFile",
				"This key is DEPRECATED and will be removed in the future",
				42,
				3,
			},
			ValidationError{
				"legal.authorsFile",
				"'https://raw.githubusercontent.com/italia/developers.italia.it/main/no_such_authors_file.txt' does not exist",
				42,
				3,
			},
		},

		// maintenance
		// maintenance.*
		"maintenance_type_missing.yml": ValidationResults{
			ValidationError{"maintenance.type", "type is a required field", 47, 3},
		},
		"maintenance_type_invalid.yml": ValidationResults{
			ValidationError{"maintenance.type", "type must be one of the following: \"internal\", \"contract\", \"community\" or \"none\"", 45, 3},
		},
		"maintenance_contacts_missing_with_type_community.yml": ValidationResults{
			ValidationError{"maintenance.contacts", "contacts is a required field when \"type\" is \"community\"", 44, 3},
		},
		"maintenance_contacts_missing_with_type_internal.yml": ValidationResults{
			ValidationError{"maintenance.contacts", "contacts is a required field when \"type\" is \"internal\"", 44, 3},
		},
		"maintenance_contacts_name_missing.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].name", "name is a required field", 0, 0},
		},
		"maintenance_contacts_email_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].email", "email must be a valid email address", 0, 0},
		},
		"maintenance_contractors_missing_with_type_contract.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "contractors is a required field when \"type\" is \"contract\"", 44, 3},
		},
		"maintenance_contractors_invalid_type.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "wrong type for this field", 47, 1},
			ValidationError{"maintenance.contractors", "contractors is a required field when \"type\" is \"contract\"", 47, 3},
		},
		"maintenance_contractors_name_missing.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].name", "name is a required field", 0, 0},
		},
		"maintenance_contractors_until_missing.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].until", "until is a required field", 0, 0},
		},
		"maintenance_contractors_until_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].until", "until must be a date with format 'YYYY-MM-DD'", 0, 0},
		},
		"maintenance_contractors_email_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].email", "email must be a valid email address", 0, 0},
		},
		"maintenance_contractors_website_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].website", "website must be an HTTP URL", 0, 0}, // TODO: line number
		},

		// localisation
		"localisation_availableLanguages_missing.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages", "availableLanguages is a required field", 50, 3},
		},
		"localisation_availableLanguages_empty.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages", "availableLanguages must contain more than 0 items", 52, 3},
		},
		"localisation_availableLanguages_invalid.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages[0]", "availableLanguages[0] must be a valid BCP 47 language", 50, 3},
		},
		// TODO: Enable this test when https://github.com/italia/publiccode-parser-go/issues/47
		// is fixed
		//
		// "localisation_availableLanguages_invalid_bcp47.yml": ValidationResults{
		// 	ValidationError{"localisation.availableLanguages[0]", "must be a valid BCP 47 language", 56, 3},
		// },
		"localisation_localisationReady_missing.yml": ValidationResults{
			ValidationError{"localisation.localisationReady", "localisationReady is a required field", 52, 3},
		},

		// dependsOn
		"dependsOn_open_name_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.name", "wrong type for this field", 56, 1},
			ValidationError{"dependsOn.open[0].name", "name is a required field", 0, 0},
		},
		"dependsOn_open_versionMin_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.versionMin", "wrong type for this field", 57, 1},
		},
		"dependsOn_open_versionMax_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.versionMax", "wrong type for this field", 57, 1},
		},
		"dependsOn_open_version_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.version", "wrong type for this field", 57, 1},
		},
		"dependsOn_open_optional_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open.optional", "wrong type for this field", 57, 1},
		},

		// it.*
		"it_countryExtensionVersion_invalid.yml": ValidationResults{
			ValidationError{"it.countryExtensionVersion", "countryExtensionVersion must be one of the following: \"0.2\" or \"1.0\"", 12, 3},
		},
		"it_riuso_codiceIPA_invalid.yml": ValidationResults{
			ValidationError{"it.riuso.codiceIPA", "codiceIPA must be a valid Italian Public Administration Code (iPA) (see https://www.indicepa.gov.it/public-services/opendata-read-service.php?dstype=FS&filename=amministrazioni.txt)", 55, 5},
		},

		// misc
		"file_encoding.yml": ValidationResults{ValidationError{"", "Invalid UTF-8", 0, 0}},
		"invalid_yaml.yml":  ValidationResults{ValidationError{"", "yaml: did not find expected key", 18, 1}},
		"mostly_empty.yml": ValidationResults{
			ValidationError{"name", "name is a required field", 1, 1},
			ValidationError{"url", "url is a required field", 1, 1},
			ValidationError{"platforms", "platforms must contain more than 0 items", 1, 1},
			ValidationError{"categories", "categories is a required field", 1, 1},
			ValidationError{"developmentStatus", "developmentStatus is a required field", 1, 1},
			ValidationError{"softwareType", "softwareType is a required field", 1, 1},
			ValidationError{"description[en-US].shortDescription", "shortDescription is a required field", 0, 0},
			ValidationError{"description[en-US].longDescription", "longDescription is a required field", 0, 0},
			ValidationError{"description[en-US].features", "features must contain more than 0 items", 0, 0},
			ValidationError{"legal.license", "license is a required field", 5, 8},
			ValidationError{"maintenance.type", "type is a required field", 6, 14},
			ValidationError{"localisation.localisationReady", "localisationReady is a required field", 4, 15},
			ValidationError{"localisation.availableLanguages", "availableLanguages is a required field", 4, 15},
		},
	}

	testFiles, _ := filepath.Glob("testdata/v0/invalid/*yml")
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

// Test v0 valid YAML testcases (testdata/v0/valid/).
func TestValidTestcasesV0(t *testing.T) {
	checkValidFiles("testdata/v0/valid/*.yml", t)
}

// Test v0 valid YAML testcases (testdata/v0/valid_with_warnings/).
func TestValidWithWarningsTestcasesV0(t *testing.T) {
	expected := map[string]error{
		"unicode_grapheme_clusters.yml": ValidationResults{
			ValidationWarning{"description.en.genericName", "This key is DEPRECATED and will be removed in the future", 23, 5},
		},
		"valid.minimal.v0.2.yml": ValidationResults{
			ValidationWarning{"publiccodeYmlVersion", "v0.2 is not the latest version, use '0.4.0'. Parsing this file as v0.4.0.", 1, 1},
		},
		"valid.minimal.v0.3.yml": ValidationResults{
			ValidationWarning{"publiccodeYmlVersion", "v0.3 is not the latest version, use '0.4.0'. Parsing this file as v0.4.0.", 1, 1},
		},
		"valid.mime_types.yml": ValidationResults{
			ValidationWarning{"inputTypes", "This key is DEPRECATED and will be removed in the future", 48, 1},
			ValidationWarning{"outputTypes", "This key is DEPRECATED and will be removed in the future", 50, 1},
		},
	}

	testFiles, _ := filepath.Glob("testdata/v0/valid_with_warnings/*yml")
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

// Test publiccode.yml remote files for key errors.
func TestDecodeValueErrorsRemote(t *testing.T) {
	testRemoteFiles := []testType{
		{"https://raw.githubusercontent.com/italia/publiccode-editor/master/publiccode.yml", ValidationResults{
			ValidationWarning{
				"publiccodeYmlVersion", "v0.2 is not the latest version, use '0.4.0'. Parsing this file as v0.4.0.", 1, 1,
			},
			ValidationWarning{"legal.authorsFile", "This key is DEPRECATED and will be removed in the future", 48, 3},
			ValidationWarning{"description.it.genericName", "This key is DEPRECATED and will be removed in the future", 12, 5},
		}},
	}

	parser, err := NewDefaultParser()
	if err != nil {
		t.Errorf("Can't create parser: %v", err)
	}

	for _, test := range testRemoteFiles {
		t.Run(fmt.Sprintf("%v", test.err), func(t *testing.T) {
			var err error

			_, err = parser.Parse(test.file)

			checkParseErrors(t, err, test)
		})
	}
}

func TestUrlMissingWithoutPath(t *testing.T) {
	expected := map[string]error{
		"url_missing.yml": ValidationResults{
			ValidationError{"url", "url is a required field", 1, 1},
		},
	}

	parser, err := NewDefaultParser()
	if err != nil {
		t.Errorf("Can't create parser: %v", err)
	}

	testFiles, _ := filepath.Glob("testdata/v0/invalid/url_missing.yml")
	for _, file := range testFiles {
		baseName := path.Base(file)
		if expected[baseName] == nil {
			t.Errorf("No expected data for file %s", baseName)
		}
		t.Run(file, func(t *testing.T) {
			_, err := parser.Parse(file)

			checkParseErrors(t, err, testType{file, expected[baseName]})
		})
	}
}

func TestIsReachable(t *testing.T) {
	parser, _ := NewParser(ParserConfig{DisableNetwork: true})

	u, _ := url.Parse("https://google.com/404")
	if reachable, _ := parser.isReachable(*u, false); !reachable {
		t.Errorf("isReachable() returned false with DisableNetwork enabled")
	}
}

// Test that the exported YAML passes validation again, and that re-exporting it
// matches the first export (lossless roundtrip).
func TestExport(t *testing.T) {
	parser, err := NewParser(ParserConfig{DisableNetwork: true})
	if err != nil {
		t.Errorf("Can't create Parser: %v", err)
	}

	publiccode, err := parser.Parse("testdata/v0/valid/valid.yml")
	if err != nil {
		t.Errorf("Failed to parse valid file: %v", err)
	}

	yaml1, err := publiccode.ToYAML()
	if err != nil {
		t.Errorf("Failed to export YAML: %v", err)
	}

	// var p2 *Parser
	// if p2, err = NewParser("/dev/null"); err != nil {
	// 	t.Errorf("Can't create Parser: %v", err)
	// }

	publiccode2, err := parser.ParseStream(bytes.NewBuffer(yaml1))
	if err != nil {
		t.Errorf("Failed to parse exported file: %v", err)
	}

	yaml2, err := publiccode2.ToYAML()
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
