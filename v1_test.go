package publiccode

import (
	"path"
	"path/filepath"
	"slices"
	"testing"
)

// Test v1 valid YAML testcases (testdata/v1/valid/).
func TestValidTestcasesV1(t *testing.T) {
	checkValidFiles("testdata/v1/valid/*.yml", t)
}

// Test v1 invalid YAML testcases (testdata/v1/invalid/).
func TestInvalidTestcasesV1(t *testing.T) {
	expected := map[string]error{
		"applicationSuite_wrong_type.yml": ValidationResults{
			ValidationError{"applicationSuite", "wrong type for this field", 4, 1},
		},
		"it_conforme.yml": ValidationResults{
			ValidationError{"IT", "unknown field \"IT\"", 118, 1},
		},
		"IT_riuso_codiceIPA.yml": ValidationResults{
			ValidationError{"IT", "unknown field \"IT\"", 118, 1},
		},
		"country_specific_section_downcase.yml": ValidationResults{
			ValidationError{"it", "unknown field \"it\"", 107, 1},
		},
		"categories_invalid.yml": ValidationResults{
			ValidationError{"categories[0]", "categories[0] must be a valid category (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/categories-list.rst)", 13, 5},
		},
		"dependsOn_open_name_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open[0]", "wrong type for this field", 56, 1},
		},
		"dependsOn_open_optional_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open[0].optional", "wrong type for this field", 57, 1},
		},
		"dependsOn_open_versionMax_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open[0].versionMax", "wrong type for this field", 57, 1},
		},
		"dependsOn_open_versionMin_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open[0].versionMin", "wrong type for this field", 57, 1},
		},
		"dependsOn_open_version_wrong_type.yml": ValidationResults{
			ValidationError{"dependsOn.open[0].version", "wrong type for this field", 57, 1},
		},
		"description_en_apiDocumentation_invalid.yml": ValidationResults{
			ValidationError{"description.en.apiDocumentation", "apiDocumentation must be an HTTP URL", 41, 5},
			ValidationError{"description.en.apiDocumentation", "'abc' not reachable: missing URL scheme", 41, 5},
		},
		"description_en_apiDocumentation_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.apiDocumentation", "wrong type for this field", 43, 1},
			ValidationError{"description", "description must contain more than 0 items", 20, 1},
		},
		"description_en_awards_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.awards", "wrong type for this field", 40, 1},
			ValidationError{"description", "description must contain more than 0 items", 18, 1},
		},
		"description_en_documentation_invalid.yml": ValidationResults{
			ValidationError{"description.en.documentation", "documentation must be an HTTP URL", 25, 5},
			ValidationError{"description.en.documentation", "'not_a_url' not reachable: missing URL scheme", 25, 5},
		},
		"description_en_documentation_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.documentation", "wrong type for this field", 25, 1},
			ValidationError{"description", "description must contain more than 0 items", 20, 1},
		},
		"description_en_features_empty.yml": ValidationResults{
			ValidationError{"description.en.features", "features must contain more than 0 items", 39, 5},
		},
		"description_en_features_missing.yml": ValidationResults{
			ValidationError{"description.en.features", "features must contain more than 0 items", 22, 5},
		},
		"description_en_gb_invalid_bcp47.yml": ValidationResults{
			ValidationError{"description", "description must be a valid BCP 47 language", 18, 1},
		},
		"description_en_genericName_too_long.yml": ValidationResults{
			ValidationError{"description.en.genericName", "unknown field \"genericName\"", 22, 1},
			ValidationError{"description", "description must contain more than 0 items", 18, 1},
		},
		"description_en_localisedName_wrong_type.yml": ValidationResults{
			ValidationError{"description.en.localisedName", "wrong type for this field", 21, 1},
			ValidationError{"description", "description must contain more than 0 items", 18, 1},
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
		"description_en_screenshots_missing_file.yml": ValidationResults{
			ValidationError{"description.en.screenshots[0]", "'no_such_file.png' is not an image: no such file: " + cwd + "/testdata/v1/invalid/no_such_file.png", 42, 9},
		},
		"description_en_shortDescription_missing.yml": ValidationResults{
			ValidationError{"description.en.shortDescription", "shortDescription is a required field", 20, 5},
		},
		"description_en_videos_invalid.yml": ValidationResults{
			ValidationError{"description.en.videos[0]", "videos[0] must be an HTTP URL", 41, 9},
			ValidationError{"description.en.videos[0]", "'ABC' is not a valid video URL supporting oEmbed: invalid oEmbed link: ABC", 41, 9},
		},
		"description_en_videos_invalid_oembed.yml": ValidationResults{
			ValidationError{"description.en.videos[0]", "'https://google.com' is not a valid video URL supporting oEmbed: invalid oEmbed link: https://google.com", 41, 9},
		},
		"description_invalid_language.yml": ValidationResults{
			ValidationError{"description", "description must be a valid BCP 47 language", 18, 1},
		},
		"developmentStatus_invalid.yml": ValidationResults{
			ValidationError{"developmentStatus", "developmentStatus must be one of the following: \"concept\", \"development\", \"beta\", \"stable\" or \"obsolete\"", 16, 1},
		},
		"developmentStatus_missing.yml": ValidationResults{
			ValidationError{"developmentStatus", "developmentStatus is a required field", 1, 1},
		},
		"developmentStatus_wrong_type.yml": ValidationResults{
			ValidationError{"developmentStatus", "wrong type for this field", 16, 1},
			ValidationError{"developmentStatus", "developmentStatus is a required field", 16, 1},
		},
		"fundedBy_uri_missing.yml": ValidationResults{
			ValidationError{"fundedBy[0].uri", "uri is a required field", 18, 5},
		},
		"fundedBy_uri_wrong_italian_pa.yml": ValidationResults{
			ValidationError{"fundedBy[0].uri", "uri is not a valid URI", 20, 5},
		},
		"fundedBy_uri_wrong_italian_pa2.yml": ValidationResults{
			ValidationError{"fundedBy[0].uri", "uri must be a valid Italian Public Administration Code (iPA) with format 'urn:x-italian-pa:[codiceIPA]' (see https://github.com/publiccodeyml/italian-organizations-ipa-vocabulary)", 19, 5},
		},
		"fundedBy_wrong_type.yml": ValidationResults{
			ValidationError{"fundedBy.name", "wrong type for this field", 18, 1},
		},
		"fundedBy_wrong_uri.yml": ValidationResults{
			ValidationError{"fundedBy[0].uri", "uri is not a valid URI", 19, 5},
		},
		"inputTypes_invalid.yml": ValidationResults{
			ValidationError{"inputTypes", "unknown field \"inputTypes\"", 14, 1},
		},
		"inputTypes_wrong_type.yml": ValidationResults{
			ValidationError{"inputTypes", "unknown field \"inputTypes\"", 14, 1},
		},
		"intendedAudience_countries_invalid_country.yml": ValidationResults{
			ValidationError{"intendedAudience.countries[2]", "countries[2] must be a valid ISO 3166-1 alpha-2 two-letter country code", 22, 7},
		},
		"intendedAudience_countries_invalid_iso_3166_1_alpha_2.yml": ValidationResults{
			ValidationError{"intendedAudience.countries[2]", "countries[2] must be a valid ISO 3166-1 alpha-2 two-letter country code", 22, 7},
		},
		"intendedAudience_countries_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.countries", "wrong type for this field", 19, 1},
		},
		"intendedAudience_scope_government.yml": ValidationResults{
			ValidationError{"intendedAudience.scope[0]", "scope[0] must be a valid scope (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/scope-list.rst)", 16, 7},
		},
		"intendedAudience_scope_invalid_scope.yml": ValidationResults{
			ValidationError{"intendedAudience.scope[0]", "scope[0] must be a valid scope (see https://github.com/publiccodeyml/publiccode.yml/blob/main/docs/standard/scope-list.rst)", 20, 9},
		},
		"intendedAudience_scope_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.scope", "wrong type for this field", 19, 1},
		},
		"intendedAudience_unsupportedCountries_invalid_country.yml": ValidationResults{
			ValidationError{"intendedAudience.unsupportedCountries[0]", "unsupportedCountries[0] must be a valid ISO 3166-1 alpha-2 two-letter country code", 20, 7},
		},
		"intendedAudience_unsupportedCountries_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience.unsupportedCountries", "wrong type for this field", 19, 1},
		},
		"intendedAudience_wrong_type.yml": ValidationResults{
			ValidationError{"intendedAudience", "wrong type for this field", 18, 1},
		},
		"isBasedOn_bad_url_array.yml": ValidationResults{
			ValidationError{"isBasedOn[1]", "isBasedOn[1] must be a valid URL", 11, 5},
		},
		"isBasedOn_bad_url_string.yml": ValidationResults{
			ValidationError{"isBasedOn[0]", "isBasedOn[0] must be a valid URL", 9, 1},
		},
		"isBasedOn_wrong_type.yml": ValidationResults{
			ValidationError{"isBasedOn.foobar", "wrong type for this field", 10, 1},
		},
		"it_countryExtensionVersion_invalid.yml": ValidationResults{
			ValidationError{"IT", "unknown field \"IT\"", 11, 1},
		},
		"it_riuso_codiceIPA_invalid.yml": ValidationResults{
			ValidationError{"IT", "unknown field \"IT\"", 53, 1},
		},
		"it_wrong_case.yml": ValidationResults{
			ValidationError{"It", "unknown field \"It\"", 107, 1},
		},
		"landingURL_invalid.yml": ValidationResults{
			ValidationError{"IT", "unknown field \"IT\"", 120, 1},
			ValidationError{"landingURL", "landingURL must be an HTTP URL", 8, 1},
			ValidationError{"landingURL", "'???' not reachable: missing URL scheme", 8, 1},
		},
		"landingURL_wrong_type.yml": ValidationResults{
			ValidationError{"landingURL", "wrong type for this field", 8, 1},
		},
		"legal_authorsFile_missing_file.yml": ValidationResults{
			ValidationError{"legal.authorsFile", "unknown field \"authorsFile\"", 42, 1},
			ValidationError{"legal.license", "license is a required field", 40, 3},
		},
		"legal_license_invalid.yml": ValidationResults{
			ValidationError{"legal.license", "license must be a valid license (see https://spdx.org/licenses)", 42, 3},
		},
		"legal_license_missing.yml": ValidationResults{
			ValidationError{"legal.license", "license is a required field", 41, 3},
		},
		"legal_missing.yml": ValidationResults{
			ValidationError{"legal.license", "license is a required field", 0, 0},
		},
		"legal_wrong_type.yml": ValidationResults{
			ValidationError{"legal", "wrong type for this field", 46, 1},
			ValidationError{"legal.license", "license is a required field", 46, 8},
		},
		"localisation_availableLanguages_empty.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages", "availableLanguages must contain more than 0 items", 52, 3},
		},
		"localisation_availableLanguages_invalid.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages[0]", "availableLanguages[0] must be a valid BCP 47 language", 53, 8},
		},
		"localisation_availableLanguages_invalid_bcp47.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages[0]", "availableLanguages[0] must be a valid BCP 47 language", 54, 8},
		},
		"localisation_availableLanguages_missing.yml": ValidationResults{
			ValidationError{"localisation.availableLanguages", "availableLanguages is a required field", 50, 3},
		},
		"localisation_localisationReady_missing.yml": ValidationResults{
			ValidationError{"localisation.localisationReady", "localisationReady is a required field", 52, 3},
		},
		"logo_absolute_path.yml": ValidationResults{
			ValidationError{"logo", "is an absolute path. Only relative paths or HTTP(s) URLs allowed", 18, 1},
		},
		"logo_file_scheme.yml": ValidationResults{
			ValidationError{"logo", "is a file:// URL. Only relative paths or HTTP(s) URLs allowed", 18, 1},
		},
		"logo_file_scheme2.yml": ValidationResults{
			ValidationError{"logo", "is a file:// URL. Only relative paths or HTTP(s) URLs allowed", 18, 1},
		},
		"logo_file_scheme3.yml": ValidationResults{
			ValidationError{"logo", "is a file:// URL. Only relative paths or HTTP(s) URLs allowed", 18, 1},
		},
		"logo_missing_file.yml": ValidationResults{
			ValidationError{"logo", "no such file: " + cwd + "/testdata/v1/invalid/no_such_file.png", 18, 1},
		},
		"logo_missing_url.yml": ValidationResults{
			ValidationError{"logo", "HTTP GET failed for https://google.com/no_such_file.png: not found", 18, 1},
		},
		"logo_unsupported_extension.yml": ValidationResults{
			ValidationError{"logo", "invalid file extension for: " + cwd + "/testdata/v1/invalid/logo.mpg", 18, 1},
		},
		"logo_wrong_type.yml": ValidationResults{
			ValidationError{"logo", "wrong type for this field", 18, 1},
		},
		"maintenance_contacts_email_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].email", "email must be a valid email address", 49, 9},
		},
		"maintenance_contacts_phone_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].phone", "phone must be a valid E.164 formatted phone number", 31, 7},
		},
		"maintenance_contacts_missing_with_type_community.yml": ValidationResults{
			ValidationError{"maintenance.contacts", "contacts is a required field when \"type\" is \"community\"", 44, 3},
		},
		"maintenance_contacts_missing_with_type_internal.yml": ValidationResults{
			ValidationError{"maintenance.contacts", "contacts is a required field when \"type\" is \"internal\"", 44, 3},
		},
		"maintenance_contacts_name_missing.yml": ValidationResults{
			ValidationError{"maintenance.contacts[0].name", "name is a required field", 47, 7},
		},
		"maintenance_contractors_email_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].email", "email must be a valid email address", 50, 8},
		},
		"maintenance_contractors_invalid_type.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "wrong type for this field", 47, 1},
			ValidationError{"maintenance.type", "type is a required field", 44, 3},
		},
		"maintenance_contractors_missing_with_type_contract.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "contractors is a required field when \"type\" is \"contract\"", 44, 3},
		},
		"maintenance_contractors_name_missing.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].name", "name is a required field", 47, 7},
		},
		"maintenance_contractors_until_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].until", "until must be a date with format 'YYYY-MM-DD'", 49, 7},
		},
		"maintenance_contractors_until_missing.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].until", "until is a required field", 47, 7},
		},
		"maintenance_contractors_website_invalid.yml": ValidationResults{
			ValidationError{"maintenance.contractors[0].website", "website must be an HTTP URL", 52, 7},
		},
		"maintenance_contractors_when_type_is_community.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "contractors must not be present unless \"type\" is \"contract\"", 46, 3},
		},
		"maintenance_contractors_when_type_is_internal.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "contractors must not be present unless \"type\" is \"contract\"", 46, 3},
		},
		"maintenance_contractors_when_type_is_none.yml": ValidationResults{
			ValidationError{"maintenance.contractors", "contractors must not be present unless \"type\" is \"contract\"", 46, 3},
		},
		"maintenance_type_invalid.yml": ValidationResults{
			ValidationError{"maintenance.type", "type must be one of the following: \"internal\", \"contract\", \"community\" or \"none\"", 45, 3},
		},
		"maintenance_type_missing.yml": ValidationResults{
			ValidationError{"maintenance.type", "type is a required field", 47, 3},
		},
		"monochromeLogo_missing_file.yml": ValidationResults{
			ValidationError{"monochromeLogo", "unknown field \"monochromeLogo\"", 18, 1},
		},
		"monochromeLogo_unsupported_extension.yml": ValidationResults{
			ValidationError{"monochromeLogo", "unknown field \"monochromeLogo\"", 18, 1},
		},
		"monochromeLogo_wrong_type.yml": ValidationResults{
			ValidationError{"monochromeLogo", "unknown field \"monochromeLogo\"", 18, 1},
		},
		"mostly_empty.yml": ValidationResults{
			ValidationError{"name", "name is a required field", 1, 1},
			ValidationError{"url", "url is a required field", 1, 1},
			ValidationError{"platforms", "platforms must contain more than 0 items", 1, 1},
			ValidationError{"developmentStatus", "developmentStatus is a required field", 1, 1},
			ValidationError{"softwareType", "softwareType is a required field", 1, 1},
			ValidationError{"description[en-US].shortDescription", "shortDescription is a required field", 3, 10},
			ValidationError{"description[en-US].longDescription", "longDescription is a required field", 3, 10},
			ValidationError{"description[en-US].features", "features must contain more than 0 items", 3, 10},
			ValidationError{"legal.license", "license is a required field", 5, 8},
			ValidationError{"maintenance.type", "type is a required field", 6, 14},
			ValidationError{"localisation.localisationReady", "localisationReady is a required field", 4, 15},
			ValidationError{"localisation.availableLanguages", "availableLanguages is a required field", 4, 15},
		},
		"name_missing.yml": ValidationResults{
			ValidationError{"name", "name is a required field", 1, 1},
		},
		"name_nil.yml": ValidationResults{
			ValidationError{"name", "name is a required field", 4, 1},
		},
		"name_wrong_type.yml": ValidationResults{
			ValidationError{"name", "wrong type for this field", 4, 1},
			ValidationError{"name", "name is a required field", 4, 1},
		},
		"organisation_uri_missing.yml": ValidationResults{
			ValidationError{"organisation.uri", "uri is a required field", 18, 3},
		},
		"organisation_uri_wrong_italian_pa.yml": ValidationResults{
			ValidationError{"organisation.uri", "uri is not a valid URI", 20, 3},
		},
		"organisation_uri_wrong_italian_pa2.yml": ValidationResults{
			ValidationError{"organisation.uri", "uri must be a valid Italian Public Administration Code (iPA) with format 'urn:x-italian-pa:[codiceIPA]' (see https://github.com/publiccodeyml/italian-organizations-ipa-vocabulary)", 19, 3},
		},
		"organisation_wrong_type.yml": ValidationResults{
			ValidationError{"organisation[0]", "wrong type for this field", 18, 1},
		},
		"organisation_wrong_uri.yml": ValidationResults{
			ValidationError{"organisation.uri", "uri is not a valid URI", 19, 3},
		},
		"outputTypes_invalid.yml": ValidationResults{
			ValidationError{"outputTypes", "unknown field \"outputTypes\"", 14, 1},
		},
		"outputTypes_wrong_type.yml": ValidationResults{
			ValidationError{"outputTypes", "unknown field \"outputTypes\"", 14, 1},
		},
		"platforms_missing.yml": ValidationResults{
			ValidationError{"platforms", "platforms must contain more than 0 items", 1, 1},
		},
		"platforms_wrong_type.yml": ValidationResults{
			ValidationError{"platforms", "wrong type for this field", 9, 1},
			ValidationError{"platforms", "platforms must contain more than 0 items", 9, 1},
		},
		"releaseDate_datetime.yml": ValidationResults{
			ValidationError{"releaseDate", "releaseDate must be a date with format 'YYYY-MM-DD'", 8, 1},
		},
		"releaseDate_empty.yml": ValidationResults{
			ValidationError{"releaseDate", "releaseDate must be a date with format 'YYYY-MM-DD'", 8, 1},
		},
		"releaseDate_invalid.yml": ValidationResults{
			ValidationError{"releaseDate", "releaseDate must be a date with format 'YYYY-MM-DD'", 8, 1},
		},
		"releaseDate_wrong_type.yml": ValidationResults{
			ValidationError{"releaseDate", "wrong type for this field", 8, 1},
		},
		"roadmap_invalid.yml": ValidationResults{
			ValidationError{"roadmap", "roadmap must be an HTTP URL", 4, 1},
			ValidationError{"roadmap", "'foobar' not reachable: missing URL scheme", 4, 1},
		},
		"roadmap_wrong_type.yml": ValidationResults{
			ValidationError{"roadmap", "wrong type for this field", 4, 1},
		},
		"softwareType_invalid.yml": ValidationResults{
			ValidationError{"softwareType", "softwareType must be one of the following: \"standalone/mobile\", \"standalone/iot\", \"standalone/desktop\", \"standalone/web\", \"standalone/backend\", \"standalone/other\", \"addon\", \"library\" or \"configurationFiles\"", 17, 1},
		},
		"softwareType_missing.yml": ValidationResults{
			ValidationError{"softwareType", "softwareType is a required field", 1, 1},
		},
		"softwareType_wrong_type.yml": ValidationResults{
			ValidationError{"softwareType", "wrong type for this field", 17, 1},
			ValidationError{"softwareType", "softwareType is a required field", 17, 1},
		},
		"softwareVersion_wrong_type.yml": ValidationResults{
			ValidationError{"softwareVersion", "wrong type for this field", 8, 1},
		},
		"unknown_field.yml": ValidationResults{
			ValidationError{"foobar", "unknown field \"foobar\"", 10, 1},
		},
		"url_invalid.yml": ValidationResults{
			ValidationError{"url", "url must be a valid URL", 6, 1},
			ValidationError{"url", "'foobar' not reachable: missing URL scheme", 6, 1},
			ValidationError{"url", "is not a valid code repository", 6, 1},
		},
		"url_missing.yml": ValidationResults{
			ValidationError{"url", "url is a required field", 1, 1},
		},
		"url_wrong_type.yml": ValidationResults{
			ValidationError{"url", "wrong type for this field", 6, 1},
			ValidationError{"url", "url is a required field", 6, 1},
		},
		"usedBy_wrong_type.yml": ValidationResults{
			ValidationError{"usedBy", "wrong type for this field", 14, 1},
		},
	}

	dir := "testdata/v1/invalid/"
	testFiles, _ := filepath.Glob(dir + "*yml")
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
	for file := range expected {
		if !slices.Contains(testFiles, dir+file) {
			t.Errorf("No expected file %s", dir+file)
		}
	}
}

// Test v1 invalid YAML testcases (testdata/v1/invalid/no-network/).
func TestInvalidTestcasesV1_NoNetwork(t *testing.T) {
	expected := map[string]error{
		"landingURL_invalid.yml": ValidationResults{
			ValidationError{"IT", "unknown field \"IT\"", 120, 1},
			ValidationError{"landingURL", "landingURL must be an HTTP URL", 8, 1},
		},
		"logo_invalid_png.yml": ValidationResults{
			ValidationError{"logo", "image: unknown format", 18, 1},
		},
		"logo_missing_file.yml": ValidationResults{
			ValidationError{"logo", "no such file: " + cwd + "/testdata/v1/invalid/no-network/no_such_file.png", 18, 1},
		},
		"monochromeLogo_invalid_png.yml": ValidationResults{
			ValidationError{"monochromeLogo", "unknown field \"monochromeLogo\"", 18, 1},
		},
	}

	dir := "testdata/v1/invalid/no-network/"
	testFiles, _ := filepath.Glob(dir + "*yml")
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
	for file := range expected {
		if !slices.Contains(testFiles, dir+file) {
			t.Errorf("No expected file %s", dir+file)
		}
	}
}
