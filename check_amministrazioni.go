package publiccode

import (
	"bufio"
	"strings"
)

// Amministrazione is an Administration from amministrazoni.txt
// Retrieved from: http://www.indicepa.gov.it/documentale/n-opendata.php
type Amministrazione struct {
	CodAmm            string
	DesAmm            string
	Comune            string
	NomeResp          string
	CognResp          string
	Cap               string
	Provincia         string
	Regione           string
	SitoIstituzionale string
	Indirizzo         string
	TitoloResp        string
	TipologiaIstat    string
	TipologiaAmm      string
	Acronimo          string
	CFValidato        string
	CF                string
	Mail1             string
	TipoMail1         string
	Mail2             string
	TipoMail2         string
	Mail3             string
	TipoMail3         string
	Mail4             string
	TipoMail4         string
	Mail5             string
	TipoMail5         string
	URLFacebook       string
	URLTwitter        string
	URLGoogleplus     string
	URLYoutube        string
	LivAccessibili    string
}

// checkCodiceIPA tells whether the codiceIPA is registered into amministrazioni.txt
// Reference: http://www.indicepa.gov.it/documentale/n-opendata.php.
func (p *Parser) checkCodiceIPA(key string, codiceiPA string) (string, error) {
	if codiceiPA == "" {
		return codiceiPA, newErrorInvalidValue(key, "empty codiceIPA key")
	}

	// Load adminisrations data from amministrazoni.txt
	dataFile, err := Asset("data/amministrazioni.txt")
	if err != nil {
		return "", err
	}
	input := string(dataFile)

	// Scan the file, line by line.
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		amm := manageLine(scanner.Text())
		// The iPA codes should be validated as case-insensitive, according
		// to the IndicePA guidelines.
		// We always fold it to lower case so that users of this library can rely
		// on a consistent case.
		if strings.EqualFold(amm.CodAmm, codiceiPA) {
			return strings.ToLower(codiceiPA), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", newErrorInvalidValue(key, "error validating Codice IPA: %s", err)
	}

	return "", newErrorInvalidValue(key, "this is not a valid Codice IPA: %s", codiceiPA)
}

// manageLine populate an Amministrazione with the values read.
func manageLine(line string) Amministrazione {
	data := strings.Split(line, "	")
	amm := Amministrazione{
		CodAmm:            data[0],
		DesAmm:            data[1],
		Comune:            data[2],
		NomeResp:          data[3],
		CognResp:          data[4],
		Cap:               data[5],
		Provincia:         data[6],
		Regione:           data[7],
		SitoIstituzionale: data[8],
		Indirizzo:         data[9],
		TitoloResp:        data[10],
		TipologiaIstat:    data[11],
		TipologiaAmm:      data[12],
		Acronimo:          data[13],
		CFValidato:        data[14],
		CF:                data[15],
		Mail1:             data[16],
		TipoMail1:         data[17],
		Mail2:             data[18],
		TipoMail2:         data[19],
		Mail3:             data[20],
		TipoMail3:         data[21],
		Mail4:             data[22],
		TipoMail4:         data[23],
		Mail5:             data[24],
		TipoMail5:         data[25],
		URLFacebook:       data[26],
		URLTwitter:        data[27],
		URLGoogleplus:     data[28],
		URLYoutube:        data[29],
		LivAccessibili:    data[30],
	}

	return amm
}
