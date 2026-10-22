package main

import "fmt"

const disclosureContext = "https://irma.app/ld/request/disclosure/v2"

// disclosureRequest is an IRMA disclosure session request. Disclose is a
// conjunction of disjunctions of conjunctions of attribute identifiers.
type disclosureRequest struct {
	Context  string       `json:"@context"`
	Disclose [][][]string `json:"disclose"`
}

// Each format takes the age limit and names one attribute that proves it.
var ageAttributeFormats = []string{
	"pbdf.pbdf.ageLimits.over%d",
	"pbdf.nijmegen.ageLimits.over%d",
	"pbdf.gemeente.personalData.over%d",
	"pbdf.pilot-amsterdam.idcard.over%d",
	"pbdf.pilot-amsterdam.passport.over%d",
}

// sessionRequests maps the "type" query parameter to its session request.
var sessionRequests = map[string]disclosureRequest{
	"12plus": ageRequest(12),
	"16plus": ageRequest(16),
	"18plus": ageRequest(18),
}

func ageRequest(limit int) disclosureRequest {
	alternatives := make([][]string, 0, len(ageAttributeFormats))
	for _, format := range ageAttributeFormats {
		alternatives = append(alternatives, []string{fmt.Sprintf(format, limit)})
	}

	return disclosureRequest{
		Context:  disclosureContext,
		Disclose: [][][]string{alternatives},
	}
}
