// Package teamsmeetingbroadcastconfiguration manages the generated meeting broadcast configuration API.
package teamsmeetingbroadcastconfiguration

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid meeting broadcast configuration input")
var ErrResponse = errors.New("unexpected meeting broadcast configuration HTTP status")

// ReadOptions selects optional Get-CsTeamsMeetingBroadcastConfiguration switches.
// ExposeSDNConfigurationJsonBlob appends ?ExposeSDNConfigurationJsonBlob=true so
// the service also returns the software defined network (SDN) provider fields.
type ReadOptions struct {
	ExposeSDNConfigurationJsonBlob bool
}

// Configuration is one meeting broadcast configuration. The Sdn* fields are only
// returned when ReadOptions.ExposeSDNConfigurationJsonBlob is set; the service
// then names the provider SdnName (Set binds SdnProviderName) and spells the
// runtime blob SdnRuntimeconfiguration. SdnAPIToken is a provider credential:
// String/GoString never print it, and RawFields/RawJSON hold the complete
// response, so callers must not log those either.
type Configuration struct {
	Identity                            *string                    `json:"Identity"`
	SupportURL                          *string                    `json:"SupportURL"`
	AllowSdnProviderForBroadcastMeeting *bool                      `json:"AllowSdnProviderForBroadcastMeeting"`
	SdnName                             *string                    `json:"SdnName"`
	SdnLicenseID                        *string                    `json:"SdnLicenseId"`
	SdnAzureSubscriptionID              *string                    `json:"SdnAzureSubscriptionId"`
	SdnAPITemplateURL                   *string                    `json:"SdnApiTemplateUrl"`
	SdnAPIToken                         *string                    `json:"SdnApiToken"`
	SdnRuntimeConfiguration             *string                    `json:"SdnRuntimeconfiguration"`
	SdnAttendeeFallbackCount            *int64                     `json:"SdnAttendeeFallbackCount"`
	RawFields                           map[string]json.RawMessage `json:"-"`
	RawJSON                             json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type plain Configuration
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Configuration(value)
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Configuration
	return json.Marshal(plain(v))
}
func (Configuration) String() string     { return "meeting broadcast configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// UpdateRequest maps Set-CsTeamsMeetingBroadcastConfiguration. Every field is
// optional: nil omits the property, a non-nil value (including false or "")
// binds it. Public 8.1.0 declares a non-nullable Boolean and plain strings, so
// no JSON null is sent. SdnAPIToken is a provider credential that may also appear
// in exposed sensitive reads; String/GoString never print it.
type UpdateRequest struct {
	AllowSdnProviderForBroadcastMeeting *bool
	SupportURL                          *string
	SdnProviderName                     *string
	SdnLicenseID                        *string
	SdnAPITemplateURL                   *string
	SdnAPIToken                         *string
	SdnRuntimeConfiguration             *string
}

func (UpdateRequest) String() string     { return "meeting broadcast configuration update (redacted)" }
func (v UpdateRequest) GoString() string { return v.String() }
