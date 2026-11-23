package migration

import "time"

// StatusOptions filters meeting migration status reads. Nil fields are not
// sent. StartTime and EndTime are formatted as the module's remoting layer
// converts a DateTime to the string query parameter ("MM/dd/yyyy HH:mm:ss",
// invariant culture, no time zone conversion).
type StatusOptions struct {
	// Identity selects one user by user principal name or SIP address.
	Identity *string
	// State is Pending, InProgress, Failed or Succeeded.
	State *string
	// MigrationType is SfbToSfb, SfbToTeams, TeamsToTeams, AllToTeams,
	// ToSameType or Unknown.
	MigrationType *string
	StartTime     *time.Time
	EndTime       *time.Time
}

// TransactionHistoryOptions filters a user's meeting migration history.
type TransactionHistoryOptions struct {
	// CorrelationID is the value returned when the migration was started.
	CorrelationID *string
	StartTime     *time.Time
	EndTime       *time.Time
}

// StartMeetingMigrationRequest starts a meeting migration for one user
// (Start-CsExMeetingMigration). Identity is required.
type StartMeetingMigrationRequest struct {
	// Identity is the user's object ID, user principal name or SIP address.
	Identity string
	// SourceMeetingType selects which meetings migrate (for example All, Teams
	// or SfB). Nil leaves the service default.
	SourceMeetingType *string
	// TargetMeetingType selects the meeting type to create (for example
	// Current or Teams). Nil leaves the service default.
	TargetMeetingType *string
}

// MoveTenantServiceInstanceRequest drives one phase of a tenant service
// instance move (Move-CsTenantServiceInstance). MoveOption is required.
type MoveTenantServiceInstanceRequest struct {
	// MoveOption is PrepForMove, StartDualSync or Finalize.
	MoveOption string
	// TargetServiceInstance is where the tenant moves. Nil is sent as an empty
	// string, as the module does when the parameter is omitted.
	TargetServiceInstance *string
}

// PartitionMoveRequest moves AVS documents between partitions
// (Move-CsAvsTenantPartition). Partition keys have the form
// "<tenantId>_<suffix>". SourcePartitionKey and TargetPartitionKey are
// required by the SDK; nil optional fields are not sent.
type PartitionMoveRequest struct {
	SourcePartitionKey    string
	TargetPartitionKey    string
	Workload              *string
	ContainerName         *string
	BasePartitionKey      *string
	NumberOfDocuments     *int64
	PercentageOfPartition *int64
	BatchSize             *int64
}
