package validators

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"golang.org/x/text/language"
)

// isBCP47StrictLanguageTag validates a BCP 47 language tag according to
// https://www.rfc-editor.org/rfc/bcp/bcp47.txt, rejecting POSIX-style tags
// (e.g. en_GB) and Unicode extensions unlike the built-in bcp47_language_tag.
//
// From https://github.com/go-playground/validator/pull/1489/.
func isBCP47StrictLanguageTag(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() == reflect.String {
		return isValidBCP47StrictLanguageTag(field.String())
	}

	//nolint:forbidigo // programming error caught at runtime, it's right to panic
	panic(fmt.Sprintf("Bad field type %s", field.Type()))
}

func isValidBCP47StrictLanguageTag(s string) bool {
	languageTagRe := regexp.MustCompile(strings.Join([]string{
		// group 1:
		`^(`,
		// irregular
		`EN-GB-OED|I-AMI|I-BNN|I-DEFAULT|I-ENOCHIAN|I-HAK|I-KLINGON|I-LUX|I-MINGO|I-NAVAJO|I-PWN|I-TAO|I-TAY|I-TSU|`,
		`SGN-BE-FR|SGN-BE-NL|SGN-CH-DE|`,
		// regular
		`ART-LOJBAN|CEL-GAULISH|NO-BOK|NO-NYN|ZH-GUOYU|ZH-HAKKA|ZH-MIN|ZH-MIN-NAN|ZH-XIANG|`,
		// privateuse
		`X-[A-Z0-9]{1,8}`,
		`)$`,

		`|`,

		// langtag
		`^`,
		`((?:[A-Z]{2,3}(?:-[A-Z]{3}){0,3})|[A-Z]{4}|[A-Z]{5,8})`, // group 2: language
		`(?:-([A-Z]{4}))?`,          // group 3: script
		`(?:-([A-Z]{2}|[0-9]{3}))?`, // group 4: region
		`(?:-((?:[A-Z0-9]{5,8}|[0-9][A-Z0-9]{3})(?:-(?:[A-Z0-9]{5,8}|[0-9][A-Z0-9]{3}))*))?`, // group 5: variant
		`(?:-((?:[A-WYZ0-9](?:-[A-Z0-9]{2,8})+)(?:-(?:[A-WYZ0-9](?:-[A-Z0-9]{2,8})+))*))?`,   // group 6: extension
		`(?:-X(?:-[A-Z0-9]{1,8})+)?`,
		`$`,
	}, ""))

	languageTag := strings.ToUpper(s)

	m := languageTagRe.FindStringSubmatch(languageTag)
	if m == nil {
		return false
	}

	grandfatheredOrPrivateuse := m[1]
	lang := m[2]
	script := m[3]
	region := m[4]
	variant := m[5]
	extension := m[6]

	if grandfatheredOrPrivateuse != "" {
		return true
	}

	switch n := len(lang); {
	case strings.Contains(lang, "-"):
		parts := strings.Split(lang, "-")

		baseLang := parts[0]

		base, err := language.ParseBase(baseLang)
		if err != nil {
			return false
		}

		if strings.ToUpper(base.String()) != baseLang {
			return false
		}

		for _, e := range parts[1:] {
			prefixes, ok := ianaExtlangs[strings.ToLower(e)]
			if !ok {
				return false
			}

			if len(prefixes) > 0 {
				found := false

				for _, p := range prefixes {
					if strings.HasPrefix(strings.ToLower(languageTag)+"-", strings.ToLower(p)) {
						found = true

						break
					}
				}

				if !found {
					return false
				}
			}
		}
	case n <= 3:
		base, err := language.ParseBase(lang)
		if err != nil {
			return false
		}

		if strings.ToUpper(base.String()) != lang {
			return false
		}
	case n == 4:
		return false
	default:
		return false
	}

	if script != "" {
		_, err := language.ParseScript(script)
		if err != nil {
			return false
		}
	}

	if region != "" {
		if len(region) == 2 {
			_, err := language.ParseRegion(region)
			if err != nil {
				return false
			}
		} else {
			if _, ok := ianaM49Codes[region]; !ok {
				return false
			}
		}
	}

	if variant != "" {
		for v := range strings.SplitSeq(variant, "-") {
			lowerVariant := strings.ToLower(v)

			_, err := language.ParseVariant(lowerVariant)
			if err != nil {
				return false
			}

			prefixes, ok := ianaVariants[lowerVariant]
			if !ok {
				return false
			}

			if len(prefixes) > 0 {
				found := false

				for _, p := range prefixes {
					if strings.HasPrefix(strings.ToLower(languageTag)+"-", strings.ToLower(p)) {
						found = true

						break
					}
				}

				if !found {
					return false
				}
			}
		}
	}

	if extension != "" {
		_, err := language.ParseExtension(extension)
		if err != nil {
			return false
		}
	}

	return true
}

// ianaVariants: variant subtags and their associated primary language prefixes.
// Source: https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
var ianaVariants = map[string][]string{
	"1606nict": {"frm"},
	"1694acad": {"fr"},
	"1901":     {"de"},
	"1959acad": {"be"},
	"1994":     {"sl-rozaj", "sl-rozaj-biske", "sl-rozaj-njiva", "sl-rozaj-osojs", "sl-rozaj-solba"},
	"1996":     {"de"},
	"abl1943":  {"pt-BR"},
	"akhmimic": {"cop"},
	"akuapem":  {"tw"},
	"alalc97":  {},
	"aluku":    {"djk"},
	"anpezo":   {"lld"},
	"ao1990":   {"pt", "gl"},
	"aranes":   {"oc"},
	"arevela":  {"hy"},
	"arevmda":  {"hy"},
	"arkaika":  {"eo"},
	"asante":   {"tw"},
	"auvern":   {"oc"},
	"baku1926": {"az", "ba", "crh", "kk", "krc", "ky", "sah", "tk", "tt", "uz"},
	"balanka":  {"blo"},
	"barla":    {"kea"},
	"basiceng": {"en"},
	"bauddha":  {"sa"},
	"bciav":    {"zbl"},
	"bcizbl":   {"zbl"},
	"biscayan": {"eu"},
	"biske":    {"sl-rozaj"},
	"blasl":    {"ase", "sgn-ase"},
	"bohairic": {"cop"},
	"bohoric":  {"sl"},
	"boont":    {"en"},
	"bornholm": {"da"},
	"cisaup":   {"oc"},
	"colb1945": {"pt"},
	"cornu":    {"en"},
	"creiss":   {"oc"},
	"dajnko":   {"sl"},
	"ekavsk":   {"sr", "sr-Latn", "sr-Cyrl"},
	"emodeng":  {"en"},
	"fascia":   {"lld"},
	"fayyumic": {"cop"},
	"fodom":    {"lld"},
	"fonipa":   {},
	"fonkirsh": {},
	"fonnapa":  {},
	"fonupa":   {},
	"fonxsamp": {},
	"gallo":    {"fr"},
	"gascon":   {"oc"},
	"gherd":    {"lld"},
	"grclass":  {"oc", "oc-aranes", "oc-auvern", "oc-cisaup", "oc-creiss", "oc-gascon", "oc-lemosin", "oc-lengadoc", "oc-nicard", "oc-provenc", "oc-vivaraup"},
	"grital":   {"oc", "oc-cisaup", "oc-nicard", "oc-provenc"},
	"grmistr":  {"oc", "oc-aranes", "oc-auvern", "oc-cisaup", "oc-creiss", "oc-gascon", "oc-lemosin", "oc-lengadoc", "oc-nicard", "oc-provenc", "oc-vivaraup"},
	"hanoi":    {"vi"},
	"hepburn":  {"ja-Latn"},
	"heploc":   {"ja-Latn-hepburn"},
	"hognorsk": {"nn"},
	"hsistemo": {"eo"},
	"huett":    {"vi"},
	"ijekavsk": {"sr", "sr-Latn", "sr-Cyrl"},
	"itihasa":  {"sa"},
	"ivanchov": {"bg"},
	"jauer":    {"rm"},
	"jyutping": {"yue"},
	"kkcor":    {"kw"},
	"kleinsch": {"kl", "kl-tunumiit"},
	"kociewie": {"pl"},
	"kscor":    {"kw"},
	"laukika":  {"sa"},
	"leidentr": {"egy"},
	"lemosin":  {"oc"},
	"lengadoc": {"oc"},
	"lipaw":    {"sl-rozaj"},
	"ltg1929":  {"ltg"},
	"ltg2007":  {"ltg"},
	"luna1918": {"ru"},
	"lycopol":  {"cop"},
	"mdcegyp":  {"egy"},
	"mdctrans": {"egy"},
	"mesokem":  {"cop"},
	"metelko":  {"sl"},
	"monoton":  {"el"},
	"ndyuka":   {"djk"},
	"nedis":    {"sl"},
	"newfound": {"en-CA"},
	"nicard":   {"oc"},
	"njiva":    {"sl-rozaj"},
	"nulik":    {"vo"},
	"osojs":    {"sl-rozaj"},
	"oxendict": {"en"},
	"pahawh2":  {"mww", "hnj"},
	"pahawh3":  {"mww", "hnj"},
	"pahawh4":  {"mww", "hnj"},
	"pamaka":   {"djk"},
	"peano":    {"la"},
	"pehoeji":  {"nan-Latn"},
	"petr1708": {"ru"},
	"pinyin":   {"zh-Latn", "bo-Latn"},
	"polyton":  {"el"},
	"provenc":  {"oc"},
	"puter":    {"rm"},
	"rigik":    {"vo"},
	"rozaj":    {"sl"},
	"rumgr":    {"rm"},
	"sahidic":  {"cop"},
	"saigon":   {"vi"},
	"scotland": {"en"},
	"scouse":   {"en"},
	"simple":   {},
	"solba":    {"sl-rozaj"},
	"sotav":    {"kea"},
	"spanglis": {"en", "es"},
	"surmiran": {"rm"},
	"sursilv":  {"rm"},
	"sutsilv":  {"rm"},
	"synnejyl": {"da"},
	"tailo":    {"nan-Latn"},
	"tarask":   {"be"},
	"tongyong": {"zh-Latn"},
	"tunumiit": {"kl"},
	"uccor":    {"kw"},
	"ucrcor":   {"kw"},
	"ulster":   {"sco"},
	"unifon":   {"en", "hup", "kyh", "tol", "yur"},
	"vaidika":  {"sa"},
	"valbadia": {"lld"},
	"valencia": {"ca"},
	"vallader": {"rm"},
	"vecdruka": {"lv"},
	"viennese": {"de"},
	"vivaraup": {"oc"},
	"wadegile": {"zh-Latn"},
	"xsistemo": {"eo"},
}

// ianaExtlangs: extended language subtags and their associated prefixes.
// Source: https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
var ianaExtlangs = map[string][]string{
	"aao": {"ar"}, "abh": {"ar"}, "abv": {"ar"}, "acm": {"ar"}, "acq": {"ar"},
	"acw": {"ar"}, "acx": {"ar"}, "acy": {"ar"}, "adf": {"ar"}, "ads": {"sgn"},
	"aeb": {"ar"}, "aec": {"ar"}, "aed": {"sgn"}, "aen": {"sgn"}, "afb": {"ar"},
	"afg": {"sgn"}, "ajp": {"ar"}, "ajs": {"sgn"}, "apc": {"ar"}, "apd": {"ar"},
	"arb": {"ar"}, "arq": {"ar"}, "ars": {"ar"}, "ary": {"ar"}, "arz": {"ar"},
	"ase": {"sgn"}, "asf": {"sgn"}, "asp": {"sgn"}, "asq": {"sgn"}, "asw": {"sgn"},
	"auz": {"ar"}, "avl": {"ar"}, "ayh": {"ar"}, "ayl": {"ar"}, "ayn": {"ar"},
	"ayp": {"ar"}, "bbz": {"ar"}, "bfi": {"sgn"}, "bfk": {"sgn"}, "bjn": {"ms"},
	"bog": {"sgn"}, "bqn": {"sgn"}, "bqy": {"sgn"}, "btj": {"ms"}, "bve": {"ms"},
	"bvl": {"sgn"}, "bvu": {"ms"}, "bzs": {"sgn"}, "cdo": {"zh"}, "cds": {"sgn"},
	"cjy": {"zh"}, "cmn": {"zh"}, "cnp": {"zh"}, "coa": {"ms"}, "cpx": {"zh"},
	"csc": {"sgn"}, "csd": {"sgn"}, "cse": {"sgn"}, "csf": {"sgn"}, "csg": {"sgn"},
	"csl": {"sgn"}, "csn": {"sgn"}, "csp": {"zh"}, "csq": {"sgn"}, "csr": {"sgn"},
	"czh": {"zh"}, "czo": {"zh"}, "doq": {"sgn"}, "dse": {"sgn"}, "dsl": {"sgn"},
	"dup": {"ms"}, "ecs": {"sgn"}, "esl": {"sgn"}, "esn": {"sgn"}, "eso": {"sgn"},
	"eth": {"sgn"}, "fcs": {"sgn"}, "fse": {"sgn"}, "fsl": {"sgn"}, "fss": {"sgn"},
	"gan": {"zh"}, "gds": {"sgn"}, "ggg": {"sgn"}, "gsg": {"sgn"}, "gsm": {"ms"},
	"gss": {"sgn"}, "gus": {"sgn"}, "hab": {"sgn"}, "haf": {"sgn"}, "hak": {"zh"},
	"hds": {"sgn"}, "hji": {"ms"}, "hks": {"sgn"}, "hos": {"sgn"}, "hps": {"sgn"},
	"hsh": {"sgn"}, "hsl": {"sgn"}, "hsn": {"zh"}, "icl": {"sgn"}, "iks": {"sgn"},
	"ils": {"sgn"}, "inl": {"sgn"}, "ins": {"sgn"}, "ise": {"sgn"}, "isg": {"sgn"},
	"isr": {"sgn"}, "jak": {"ms"}, "jax": {"ms"}, "jcs": {"sgn"}, "jhs": {"sgn"},
	"jls": {"sgn"}, "jos": {"sgn"}, "jsl": {"sgn"}, "jus": {"sgn"}, "kgi": {"sgn"},
	"knn": {"ms"}, "kvb": {"ms"}, "kvk": {"sgn"}, "kvr": {"ms"}, "kxd": {"ms"},
	"lbs": {"sgn"}, "lce": {"ms"}, "lcf": {"ms"}, "liw": {"ms"}, "lls": {"sgn"},
	"lsg": {"sgn"}, "lsl": {"sgn"}, "lso": {"sgn"}, "lsp": {"sgn"}, "lst": {"sgn"},
	"lsy": {"sgn"}, "ltg": {"lv"}, "lvs": {"lv"}, "lws": {"sgn"}, "lzh": {"zh"},
	"max": {"ms"}, "mdl": {"sgn"}, "meo": {"ms"}, "mfa": {"ms"}, "mfb": {"ms"},
	"mfs": {"sgn"}, "min": {"ms"}, "mnp": {"zh"}, "mqg": {"ms"}, "mre": {"sgn"},
	"msd": {"sgn"}, "msi": {"ms"}, "msr": {"sgn"}, "mui": {"ms"}, "mzc": {"sgn"},
	"mzg": {"sgn"}, "mzy": {"sgn"}, "nan": {"zh"}, "nbs": {"sgn"}, "ncs": {"sgn"},
	"nsi": {"sgn"}, "nsl": {"sgn"}, "nsp": {"sgn"}, "nsr": {"sgn"}, "nzs": {"sgn"},
	"okl": {"sgn"}, "orn": {"ms"}, "ors": {"ms"}, "pel": {"ms"}, "pga": {"ar"},
	"pgz": {"sgn"}, "pks": {"sgn"}, "prl": {"sgn"}, "prz": {"sgn"}, "psc": {"sgn"},
	"psd": {"sgn"}, "pse": {"ms"}, "psg": {"sgn"}, "psl": {"sgn"}, "pso": {"sgn"},
	"psp": {"sgn"}, "psr": {"sgn"}, "pys": {"sgn"}, "rms": {"sgn"}, "rsi": {"sgn"},
	"rsl": {"sgn"}, "rsm": {"sgn"}, "sdl": {"sgn"}, "sfb": {"sgn"}, "sfs": {"sgn"},
	"sgg": {"sgn"}, "sgx": {"sgn"}, "shu": {"ar"}, "slf": {"sgn"}, "sls": {"sgn"},
	"sqs": {"sgn"}, "sqx": {"sgn"}, "ssh": {"ar"}, "ssp": {"sgn"}, "ssr": {"sgn"},
	"svk": {"sgn"}, "swc": {"sw"}, "swh": {"sw"}, "swl": {"sgn"}, "syy": {"sgn"},
	"szs": {"sgn"}, "tmw": {"ms"}, "tse": {"sgn"}, "tsm": {"sgn"}, "tsq": {"sgn"},
	"tss": {"sgn"}, "tsy": {"sgn"}, "tza": {"ms"}, "ugn": {"sgn"}, "ugy": {"sgn"},
	"ukl": {"sgn"}, "uks": {"sgn"}, "urk": {"ms"}, "uzn": {"uz"}, "uzs": {"uz"},
	"vgt": {"sgn"}, "vkk": {"ms"}, "vkt": {"ms"}, "vsi": {"sgn"}, "vsl": {"sgn"},
	"vsv": {"sgn"}, "wbs": {"sgn"}, "wuu": {"zh"}, "xki": {"sgn"}, "xml": {"sgn"},
	"xmm": {"ms"}, "xms": {"sgn"}, "yds": {"sgn"}, "ygs": {"sgn"}, "yhs": {"sgn"},
	"ysl": {"sgn"}, "yss": {"sgn"}, "zib": {"sgn"}, "zlm": {"ms"}, "zmi": {"ms"},
	"zsl": {"sgn"}, "zsm": {"ms"},
}

// ianaM49Codes: UN M.49 region codes present in the IANA subtag registry.
// Source: https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
var ianaM49Codes = map[string]struct{}{
	"001": {}, "002": {}, "003": {}, "005": {}, "009": {},
	"011": {}, "013": {}, "014": {}, "015": {}, "017": {},
	"018": {}, "019": {}, "021": {}, "029": {}, "030": {},
	"034": {}, "035": {}, "039": {}, "053": {}, "054": {},
	"057": {}, "061": {}, "142": {}, "143": {}, "145": {},
	"150": {}, "151": {}, "154": {}, "155": {}, "202": {},
	"419": {},
}
