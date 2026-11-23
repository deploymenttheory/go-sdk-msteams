package provisioningdiagnostics

import "encoding/json"

// DeploymentRegion is a provisioning region of the OCE user-history route.
type DeploymentRegion string

// Regions accepted by Get-CsUserProvHistory.
const (
	RegionNOAM DeploymentRegion = "NOAM"
	RegionEMEA DeploymentRegion = "EMEA"
	RegionAPAC DeploymentRegion = "APAC"
)

// DirectoryDeployment is a directory sync deployment name. The OCE directory
// sync body sends its index in this list as a number; the runHandler route
// sends its name.
type DirectoryDeployment string

// Deployments in the module's DirectoryObjectSyncDeployment order.
const (
	DeploymentLocal         DirectoryDeployment = "Local"
	DeploymentINT           DirectoryDeployment = "INT"
	DeploymentCentralUS     DirectoryDeployment = "CentralUS"
	DeploymentEastUS        DirectoryDeployment = "EastUS"
	DeploymentNorthEurope   DirectoryDeployment = "NorthEurope"
	DeploymentWestEurope    DirectoryDeployment = "WestEurope"
	DeploymentJapanWest     DirectoryDeployment = "JapanWest"
	DeploymentSouthEastAsia DirectoryDeployment = "SouthEastAsia"
	DeploymentGCCMCentralUS DirectoryDeployment = "GCCMCentralUS"
	DeploymentGCCMEastUS    DirectoryDeployment = "GCCMEastUS"
	DeploymentUSGovArizona  DirectoryDeployment = "USGovArizona"
	DeploymentUSGovTexas    DirectoryDeployment = "USGovTexas"
	DeploymentUSNatWest     DirectoryDeployment = "USNatWest"
	DeploymentUSNatEast     DirectoryDeployment = "USNatEast"
	DeploymentUSSecWest     DirectoryDeployment = "USSecWest"
	DeploymentUSSecEast     DirectoryDeployment = "USSecEast"
	DeploymentUSDoDCentral  DirectoryDeployment = "USDoDCentral"
	DeploymentUSDoDEast     DirectoryDeployment = "USDoDEast"
	DeploymentChinaEast3    DirectoryDeployment = "ChinaEast3"
	DeploymentChinaNorth3   DirectoryDeployment = "ChinaNorth3"
)

var deployments = []DirectoryDeployment{DeploymentLocal, DeploymentINT, DeploymentCentralUS, DeploymentEastUS, DeploymentNorthEurope,
	DeploymentWestEurope, DeploymentJapanWest, DeploymentSouthEastAsia, DeploymentGCCMCentralUS, DeploymentGCCMEastUS, DeploymentUSGovArizona,
	DeploymentUSGovTexas, DeploymentUSNatWest, DeploymentUSNatEast, DeploymentUSSecWest, DeploymentUSSecEast, DeploymentUSDoDCentral,
	DeploymentUSDoDEast, DeploymentChinaEast3, DeploymentChinaNorth3}

// DirectoryObjectClass is the object class of an OCE directory sync, sent as
// its index in this list.
type DirectoryObjectClass string

// Object classes in the module's DirectoryObjectClass order.
const (
	ObjectClassCompany        DirectoryObjectClass = "Company"
	ObjectClassContact        DirectoryObjectClass = "Contact"
	ObjectClassGroup          DirectoryObjectClass = "Group"
	ObjectClassUser           DirectoryObjectClass = "User"
	ObjectClassSubscribedPlan DirectoryObjectClass = "SubscribedPlan"
)

var objectClasses = []DirectoryObjectClass{ObjectClassCompany, ObjectClassContact, ObjectClassGroup, ObjectClassUser, ObjectClassSubscribedPlan}

func index[T comparable](values []T, value T) (int, bool) {
	for i, candidate := range values {
		if candidate == value {
			return i, true
		}
	}
	return 0, false
}

// DirectoryObjectSyncRequest is Invoke-CsDirectoryObjectSync. TenantID,
// Deployment, ServiceInstance (MicrosoftCommunicationsOnline/…) and ObjectClass
// are required. A non-empty ObjectIDs is sent instead of ObjectID.
type DirectoryObjectSyncRequest struct {
	TenantID                        string
	Deployment                      DirectoryDeployment
	ServiceInstance                 string
	ObjectClass                     DirectoryObjectClass
	ObjectID                        string
	ObjectIDs                       []string
	SynchronizeTenantWithAllObjects bool
	IsValidationRequest             bool
	ScenariosToSuppress             *string
}

// NgtProvHandlerRequest is Invoke-CsGenericNgtProvCommand. Payload is sent as
// the JSON request body.
type NgtProvHandlerRequest struct {
	TenantID        string
	Region          DirectoryDeployment
	HandlerName     string
	Target          string
	ServiceInstance string
	Payload         json.RawMessage
}

// UserHistory is a user's provisioning history (Get-CsUserProvHistory).
type UserHistory struct {
	UserID         *string                    `json:"UserId"`
	TenantID       *string                    `json:"TenantId"`
	VersionHistory []*VersionInformation      `json:"VersionHistory"`
	RawFields      map[string]json.RawMessage `json:"-"`
}

func (v *UserHistory) UnmarshalJSON(raw []byte) error {
	type plain UserHistory
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = UserHistory(value)
	return nil
}

// VersionInformation is one provisioned version and its orders.
type VersionInformation struct {
	Version            *int64                     `json:"Version"`
	VersionCreatedTime *string                    `json:"VersionCreatedTime"`
	OrderHistory       []*OrderInformation        `json:"OrderHistory"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *VersionInformation) UnmarshalJSON(raw []byte) error {
	type plain VersionInformation
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = VersionInformation(value)
	return nil
}

// OrderInformation is one provisioning order. Times keep the service's text.
type OrderInformation struct {
	OrderID       *string                    `json:"OrderId"`
	Faulted       *bool                      `json:"Faulted"`
	ProcessorType *string                    `json:"ProcessorType"`
	OrderStatus   *string                    `json:"OrderStatus"`
	DueDate       *string                    `json:"DueDate"`
	TimeQueued    *string                    `json:"TimeQueued"`
	TimePickedUp  *string                    `json:"TimePickedUp"`
	TimeCompleted *string                    `json:"TimeCompleted"`
	AttemptCount  *int64                     `json:"AttemptCount"`
	JarvisURI     *string                    `json:"JarvisUri"`
	RawFields     map[string]json.RawMessage `json:"-"`
}

func (v *OrderInformation) UnmarshalJSON(raw []byte) error {
	type plain OrderInformation
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = OrderInformation(value)
	return nil
}
