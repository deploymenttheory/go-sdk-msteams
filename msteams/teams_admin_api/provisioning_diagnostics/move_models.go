package provisioningdiagnostics

import "encoding/json"

// UserMove identifies the user a hosted-migration step acts on. TenantID is
// sent as X-MS-Tenant-Id, as the module sends the connected tenant.
type UserMove struct {
	TenantID   string
	UserSipURI string
}

// TeamsData carries the Teams data checks of a hosted-migration step. The
// module always sends all three flags.
type TeamsData struct {
	MoveToTeams          bool `json:"MoveToTeams"`
	CheckEnterpriseVoice bool `json:"CheckEnterpriseVoice"`
	CheckCpc             bool `json:"CheckCpc"`
}

// LocalDeployment describes the on-premises deployment the user moves from.
// An empty HostingProviderFqdn is omitted.
type LocalDeployment struct {
	PresenceFqdn        string `json:"PresenceFqdn"`
	RegistrarFqdn       string `json:"RegistrarFqdn"`
	HostingProviderFqdn string `json:"HostingProviderFqdn,omitempty"`
	MajorVersion        int32  `json:"MajorVersion"`
}

// ValidateUserMoveRequest is Move-CsInternalHelper -ActionType Validateuser
// (parameter set Validate).
type ValidateUserMoveRequest struct {
	UserMove
	CmdletVersion   string
	LocalDeployment LocalDeployment
	TeamsData       TeamsData
	MoveToCloud     bool
	Force           bool
}

// UserMoveVersionRequest is Move-CsInternalHelper -ActionType Beginmove or
// Completemove (parameter set BeginAndCompleteMove).
type UserMoveVersionRequest struct {
	UserMove
	MajorVersion string
}

// RehomeUserMoveRequest is Move-CsInternalHelper -ActionType Rehomeuser
// (parameter set Rehome).
type RehomeUserMoveRequest struct {
	UserMove
	TeamsData   TeamsData
	MoveToCloud bool
}

// MovedResourceDataRequest is Move-CsInternalHelper -ActionType
// Setmovedresourcedata (parameter set MoveResourcedata).
type MovedResourceDataRequest struct {
	UserMove
	ResourceData string
	MajorVersion string
	TeamsData    TeamsData
}

// DeploymentInfo is the deployment a validated user is homed on.
type DeploymentInfo struct {
	PresenceFqdn        *string                    `json:"PresenceFqdn"`
	RegistrarFqdn       *string                    `json:"RegistrarFqdn"`
	HostingProviderFqdn *string                    `json:"HostingProviderFqdn"`
	MajorVersion        *int32                     `json:"MajorVersion"`
	RawFields           map[string]json.RawMessage `json:"-"`
}

func (v *DeploymentInfo) UnmarshalJSON(raw []byte) error {
	type plain DeploymentInfo
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = DeploymentInfo(value)
	return nil
}

// MigrationData is opaque move state returned by BeginUserMove.
type MigrationData struct {
	Datastr   *string                    `json:"Datastr"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *MigrationData) UnmarshalJSON(raw []byte) error {
	type plain MigrationData
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = MigrationData(value)
	return nil
}

// KeyValue is one policy entry returned by GetUserMovePolicies.
type KeyValue struct {
	Key       *string                    `json:"Key"`
	Value     *string                    `json:"Value"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *KeyValue) UnmarshalJSON(raw []byte) error {
	type plain KeyValue
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = KeyValue(value)
	return nil
}
