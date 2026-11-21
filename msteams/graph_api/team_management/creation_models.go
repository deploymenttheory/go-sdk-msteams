package teammanagement

import (
	"errors"
	"time"

	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

var (
	ErrMailNicknameLimit      = errors.New("generated mail nickname candidates exhausted")
	ErrCreationReconciliation = errors.New("created group requires further reconciliation before cleanup")
)

// GroupTemplate selects the pinned group provisioning presets. These are not
// native POST /teams template IDs, nor tenant eligibility guarantees.
type GroupTemplate string

const (
	OrdinaryGroup  GroupTemplate = ""
	EducationClass GroupTemplate = "EDU_Class"
	EducationPLC   GroupTemplate = "EDU_PLC"
)

// CreationOptions bounds the complete foreground workflow and each received-404
// propagation sequence. Zero selects five minutes, four attempts per stage
// (initial plus three retries) and ten seconds between attempts. Cleanup has a
// separate thirty-second default.
// Cleanup may continue after caller cancellation, within that independent bound.
type CreationOptions struct {
	Timeout        time.Duration
	MaxAttempts    int
	Interval       time.Duration
	CleanupTimeout time.Duration
}

// CreateRequest is New-Team/CreateTeam's group-first input. Nil MailNickname
// generates/checks msteams_ plus six hex characters; an explicit alias is never
// normalized or substituted. Empty Owner selects /me only in delegated mode.
type CreateRequest struct {
	DisplayName        string
	MailNickname       *string
	Description        *string
	Classification     *string
	Visibility         *group.Visibility
	Owner              string
	Template           GroupTemplate
	Settings           team.Settings
	RetainCreatedGroup bool
	Options            CreationOptions
}

type CleanupDisposition string

const (
	CleanupNotRequired       CleanupDisposition = "not-required"
	CleanupSuppliedGroup     CleanupDisposition = "supplied-group-retained"
	CleanupRetainedByRequest CleanupDisposition = "retained-by-request"
	CleanupTeamExists        CleanupDisposition = "team-exists-retained"
	CleanupUnresolved        CleanupDisposition = "unresolved-retained"
	CleanupDeleted           CleanupDisposition = "group-deleted"
	CleanupDeleteUnconfirmed CleanupDisposition = "delete-unconfirmed"
)

// CreationCleanup distinguishes observed existence, requested retention, unknown
// mutation outcomes and a confirmed 204 delete. A failed read is never absence.
type CreationCleanup struct {
	Disposition            CleanupDisposition
	Team                   *team.Team
	ReconciliationResponse *resty.Response
	Response               *resty.Response
	Error                  error
}

func (CreationCleanup) String() string     { return "team creation cleanup (redacted)" }
func (c CreationCleanup) GoString() string { return c.String() }

// CreationResult retains known IDs and completed stages on every failure.
// CreatedGroupOwned is established only by this call's successful group create;
// it is never inferred from an input ID. A nil Value does not undo earlier writes.
// PrimaryError and Cleanup.Error remain separate; the returned error joins them.
type CreationResult struct {
	Result
	OwnerID               string
	OwnerSource           directoryuser.Source
	MailNickname          string
	GeneratedMailNickname bool
	CreatedGroupOwned     bool
	CreatedGroup          *group.Group
	CreatedGroupResponse  *resty.Response
	Teamified             *team.Team
	PrimaryError          error
	PrimaryResponse       *resty.Response
	Cleanup               CreationCleanup
	teamificationRejected bool
}

func (CreationResult) String() string     { return "team creation result (redacted)" }
func (r CreationResult) GoString() string { return r.String() }
