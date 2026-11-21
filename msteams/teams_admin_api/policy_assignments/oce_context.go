package policyassignments

import (
	"regexp"
	"strings"
)

// OCEEnvironment identifies the pinned module's conditional service profile.
// It never changes a session's origin, authority or credential audience.
type OCEEnvironment string

const (
	OCEMultiTenant OCEEnvironment = "MultiTenant"
	OCEIntegration OCEEnvironment = "INT"
	OCEItar        OCEEnvironment = "Itar"
	OCEGCCH        OCEEnvironment = "GCCH"
	OCEDOD         OCEEnvironment = "DOD"
	OCEAg08        OCEEnvironment = "Ag08"
	OCEAg09        OCEEnvironment = "Ag09"
	OCEGallatin    OCEEnvironment = "Gallatin"
)

type Region string

const (
	RegionINT      Region = "INT"
	RegionNOAM     Region = "NOAM"
	RegionEMEA     Region = "EMEA"
	RegionAPAC     Region = "APAC"
	RegionUSSEC    Region = "ussec"
	RegionUSNAT    Region = "usnat"
	RegionGCCH     Region = "gcch"
	RegionDOD      Region = "dod"
	RegionGallatin Region = "Gallatin"
)

// OCEContext explicitly supplies the target tenant. Empty Environment selects
// the profile matching the session's configured origin. Explicit profiles must
// match that same origin; these options cannot forward a token to another cloud.
// Conditional OCE eligibility remains independent of ordinary tenant roles.
type OCEContext struct {
	TenantID    string
	Environment OCEEnvironment
}

func (OCEContext) String() string     { return "OCE assignment context (redacted)" }
func (v OCEContext) GoString() string { return v.String() }

var assignmentGUID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func canonicalGUID(value string) (string, bool) {
	if !assignmentGUID.MatchString(value) {
		return "", false
	}
	return strings.ToLower(value), true
}

func normalizeRegion(value Region) (Region, bool) {
	switch strings.ToLower(string(value)) {
	case "int":
		return RegionINT, true
	case "noam":
		return RegionNOAM, true
	case "emea":
		return RegionEMEA, true
	case "apac":
		return RegionAPAC, true
	case "ussec":
		return RegionUSSEC, true
	case "usnat":
		return RegionUSNAT, true
	case "gcch":
		return RegionGCCH, true
	case "dod":
		return RegionDOD, true
	case "gallatin":
		return RegionGallatin, true
	default:
		return "", false
	}
}

type oceProfile struct {
	environment OCEEnvironment
	origin      string
	regions     []Region
}

func oceProfiles() []oceProfile {
	return []oceProfile{
		{OCEMultiTenant, "https://api.interfaces.records.teams.microsoft.com", []Region{RegionNOAM, RegionEMEA, RegionAPAC}},
		{OCEIntegration, "https://dev.api.interfaces.records.teams.microsoft.com", []Region{RegionINT}},
		{OCEGCCH, "https://api.interfaces.records.gov.teams.microsoft.us", []Region{RegionGCCH}},
		{OCEItar, "https://api.interfaces.records.gov.teams.microsoft.us", []Region{RegionGCCH}},
		{OCEDOD, "https://api.interfaces.records.dod.teams.microsoft.us", []Region{RegionDOD}},
		{OCEAg08, "https://api.interfaces.records.teams.eaglex.ic.gov", []Region{RegionUSNAT}},
		{OCEAg09, "https://api.interfaces.records.teams.microsoft.scloud", []Region{RegionUSSEC}},
		{OCEGallatin, "https://api.interfaces.records.teams.microsoftonline.cn", []Region{RegionGallatin}},
	}
}
