package extintegration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	shifts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/shifts_connections"
)

// These authored source-contract vectors test identity handoff, not WFM access,
// accepted-operation completion, persisted mappings or live cleanup qualification.
const shiftsV2 = "/Teams.ShiftsConnector/api/v2"
const shiftsV1 = "/Teams.ShiftsConnector/api/v1"
const connectionID = "00000000-0000-4000-8000-000000000011"
const instanceID = "00000000-0000-4000-8000-000000000012"
const mapOperationID = "00000000-0000-4000-8000-000000000013"
const shiftsTeamID = "00000000-0000-4000-8000-000000000014"
const createConnectionBody = `{"connectorId":"connector","connectorSpecificSettings":{},"name":"owned-connection","state":"Inactive"}`
const connectionReply = `{"id":"` + connectionID + `","name":"owned-connection","connectorId":"connector"}`
const createInstanceBody = `{"connectionId":"` + connectionID + `","connectorAdminEmails":[],"name":"owned-instance","syncFrequencyInMin":0}`
const instanceReply = `{"id":"` + instanceID + `","connectionId":"` + connectionID + `","connectorAdminEmails":[],"syncFrequencyInMin":0}`
const mappingBody = `{"teamMappings":[{"teamId":"` + shiftsTeamID + `","wfmTeamId":"wfm-team"}]}`

func createShiftsPair(ctx context.Context, sdk *msteams.Client) (*shifts.Connection, *shifts.Instance, error) {
	s := sdk.TeamsAdminAPI.ShiftsConnections
	connection, _, err := s.CreateConnection(ctx, shifts.CreateConnectionRequest{Name: "owned-connection", ConnectorID: "connector", State: "Inactive", ConnectorSpecificSettings: shifts.Settings{}})
	if err != nil {
		return connection, nil, err
	}
	instance, _, err := s.CreateInstance(ctx, shifts.InstanceRequest{ConnectionID: connection.ID, Name: ptr("owned-instance"), ConnectorAdminEmails: []string{}, SyncFrequencyInMin: ptr(int32(0))})
	return connection, instance, err
}
func shiftsCreateSteps() []testutil.Exchange {
	return []testutil.Exchange{
		exchange(false, "POST", shiftsV2+"/connections", "New-CsTeamsShiftsConnection", createConnectionBody, 201, connectionReply),
		exchange(false, "POST", shiftsV2+"/connectorinstances", "New-CsTeamsShiftsConnectionInstance", createInstanceBody, 201, instanceReply),
	}
}

func TestUnitShiftsConnectionInstanceAndAsyncOperationKeepDistinctIDs(t *testing.T) {
	t.Parallel()
	steps := append(shiftsCreateSteps(),
		exchange(false, "PUT", shiftsV1+"/connectorinstances/"+instanceID+"/teamconnects", "New-CsTeamsShiftsConnectionBatchTeamMap", mappingBody, 202, `{"operationId":"`+mapOperationID+`","status":"NotStarted","lastActionDateTime":null}`),
		exchange(false, "GET", shiftsV1+"/operations/"+mapOperationID, "Get-CsTeamsShiftsConnectionOperation", "", 200, `{"Id":"`+mapOperationID+`","wfmConnectorInstanceId":"`+instanceID+`","status":"Completed","details":[{"status":"Failed","error":"WfmTeamNotFound","properties":{"teamId":"`+shiftsTeamID+`"}}]}`),
		exchange(false, "GET", shiftsV1+"/connectorinstances/"+instanceID+"/teamconnects", "Get-CsTeamsShiftsConnectionTeamMap", "", 200, `{"teamConnects":[]}`),
	)
	run(t, "shifts-returned-identity-handoff", steps, nil, func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		connection, instance, err := createShiftsPair(ctx, sdk)
		if err != nil {
			return testutil.Observation{Err: err}
		}
		if connection.ID == nil || *connection.ID != connectionID || instance.ID == nil || *instance.ID != instanceID || instance.ConnectionID == nil || *instance.ConnectionID != *connection.ID || instance.SyncFrequencyInMin == nil || *instance.SyncFrequencyInMin != 0 || instance.ConnectorAdminEmails == nil {
			t.Fatal("connection/instance identity or explicit empty/zero was lost")
		}
		s := sdk.TeamsAdminAPI.ShiftsConnections
		accepted, response, err := s.MapTeams(ctx, *instance.ID, []shifts.TeamMapping{{TeamID: shiftsTeamID, WfmTeamID: "wfm-team"}})
		if err != nil {
			return testutil.Observation{Err: err}
		}
		if response == nil || response.StatusCode() != 202 || accepted.OperationID == nil || *accepted.OperationID != mapOperationID || accepted.Status == nil || *accepted.Status != "NotStarted" {
			t.Fatal("accepted async receipt lost its distinct operation ID")
		}
		operation, response, err := s.GetOperationByID(ctx, *accepted.OperationID)
		if err != nil {
			return testutil.Observation{Err: err}
		}
		// A completed envelope can still contain failed details. Neither 202 nor the
		// outer Completed string alone proves a successfully persisted team mapping.
		if response == nil || response.StatusCode() != 200 || operation.ID == nil || *operation.ID != *accepted.OperationID || operation.WfmConnectorInstanceID == nil || *operation.WfmConnectorInstanceID != *instance.ID || len(operation.Details) != 1 || operation.Details[0].Status == nil || *operation.Details[0].Status != "Failed" || operation.Details[0].Error == nil || *operation.Details[0].Error != "WfmTeamNotFound" {
			t.Fatal("operation handoff or per-detail failure was flattened")
		}
		maps, response, err := s.ListTeamMaps(ctx, *instance.ID)
		if err == nil && (response == nil || response.StatusCode() != 200 || maps == nil || len(maps) != 0) {
			t.Fatal("failed operation invented a persisted map")
		}
		return testutil.Observation{Err: err}
	})
}

func TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup(t *testing.T) {
	t.Parallel()
	steps := append(shiftsCreateSteps(),
		exchange(false, "PUT", shiftsV1+"/connectorinstances/"+instanceID+"/teamconnects", "New-CsTeamsShiftsConnectionBatchTeamMap", mappingBody, 403, `{"code":"Forbidden"}`),
		exchange(false, "DELETE", shiftsV2+"/connectorinstances/"+instanceID, "Remove-CsTeamsShiftsConnectionInstance", "", 204, ""),
		exchange(false, "GET", shiftsV2+"/connectorinstances/"+instanceID, "Get-CsTeamsShiftsConnectionInstance", "", 404, `{"code":"NotFound"}`),
		exchange(false, "DELETE", shiftsV2+"/connections/"+connectionID, "Remove-CsTeamsShiftsConnection", "", 204, ""),
		exchange(false, "GET", shiftsV2+"/connections/"+connectionID, "Get-CsTeamsShiftsConnection", "", 404, `{"code":"NotFound"}`),
	)
	run(t, "shifts-sync-denial-independent-cleanup", steps, auth.ErrForbidden, func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		connection, instance, err := createShiftsPair(ctx, sdk)
		if err != nil {
			return testutil.Observation{Err: err}
		}
		if connection.ID == nil || instance.ID == nil {
			t.Fatal("successful parent receipts lost their IDs")
		}
		s := sdk.TeamsAdminAPI.ShiftsConnections
		operation, response, operationErr := s.MapTeams(ctx, *instance.ID, []shifts.TeamMapping{{TeamID: shiftsTeamID, WfmTeamID: "wfm-team"}})
		if !errors.Is(operationErr, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 || operation != nil {
			t.Fatal("denied sync invented an operation or lost its receipt")
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		// Cleanup has its own bounded parent context even when the operation caller
		// has canceled. This test controls exact owned receipts; it is not a live journal.
		cleanup := context.WithoutCancel(canceled)
		cleanup, stop := context.WithTimeout(cleanup, time.Minute)
		defer stop()
		if _, err := s.DeleteInstanceByID(cleanup, *instance.ID); err != nil {
			t.Fatal("independent instance deletion failed")
		}
		var absent *client.APIError
		if _, response, err := s.GetInstanceByID(cleanup, *instance.ID); !errors.As(err, &absent) || absent.StatusCode != 404 || response == nil || response.StatusCode() != 404 {
			t.Fatal("instance absence was not independently read")
		}
		if _, err := s.DeleteConnectionByID(cleanup, *connection.ID); err != nil {
			t.Fatal("independent connection deletion failed")
		}
		absent = nil
		if _, response, err := s.GetConnectionByID(cleanup, *connection.ID); !errors.As(err, &absent) || absent.StatusCode != 404 || response == nil || response.StatusCode() != 404 {
			t.Fatal("connection absence was not independently read")
		}
		return testutil.Observation{Err: operationErr}
	})
}
