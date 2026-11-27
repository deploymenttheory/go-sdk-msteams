package teamsclientconfiguration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

type modelCase struct {
	Name, Input string
	Observation struct {
		Failure *string
		Value   json.RawMessage
	}
}

func modelCases(t *testing.T) []modelCase {
	t.Helper()

	fixture := struct{ Models []modelCase }{
		Models: []modelCase{
			{
				Name:  "constructor",
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "all-settings",
				Input: "{\"AllowEmailIntoChannel\":false,\"RestrictedSenderList\":\"Future-RestrictedSenderList\",\"AllowDropBox\":false,\"AllowBox\":false,\"AllowGoogleDrive\":false,\"AllowShareFile\":false,\"AllowEgnyte\":false,\"AllowOrganizationTab\":false,\"AllowSkypeBusinessInterop\":false,\"ContentPin\":\"Future-ContentPin\",\"AllowResourceAccountSendMessage\":false,\"ResourceAccountContentAccess\":\"Future-ResourceAccountContentAccess\",\"AllowGuestUser\":false,\"AllowScopedPeopleSearchandAccess\":false,\"AllowRoleBasedChatPermissions\":false,\"ExtendedWorkInfoInPeopleSearch\":false,\"UseUnifiedDomain\":\"Future-UseUnifiedDomain\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":false,\"AllowDropBox\":false,\"AllowEgnyte\":false,\"AllowEmailIntoChannel\":false,\"AllowGoogleDrive\":false,\"AllowGuestUser\":false,\"AllowOrganizationTab\":false,\"AllowResourceAccountSendMessage\":false,\"AllowRoleBasedChatPermissions\":false,\"AllowScopedPeopleSearchandAccess\":false,\"AllowShareFile\":false,\"AllowSkypeBusinessInterop\":false,\"ContentPin\":\"Future-ContentPin\",\"ExtendedWorkInfoInPeopleSearch\":false,\"Identity\":null,\"ResourceAccountContentAccess\":\"Future-ResourceAccountContentAccess\",\"RestrictedSenderList\":\"Future-RestrictedSenderList\",\"UseUnifiedDomain\":\"Future-UseUnifiedDomain\"}"),
				},
			},
			{
				Name:  "AllowEmailIntoChannel-null",
				Input: "{\"AllowEmailIntoChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowEmailIntoChannel-false",
				Input: "{\"AllowEmailIntoChannel\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":false,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowEmailIntoChannel-true",
				Input: "{\"AllowEmailIntoChannel\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":true,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "RestrictedSenderList-null",
				Input: "{\"RestrictedSenderList\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "RestrictedSenderList-empty",
				Input: "{\"RestrictedSenderList\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":\"\",\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "RestrictedSenderList-future",
				Input: "{\"RestrictedSenderList\":\"Future-RestrictedSenderList\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":\"Future-RestrictedSenderList\",\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowDropBox-null",
				Input: "{\"AllowDropBox\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowDropBox-false",
				Input: "{\"AllowDropBox\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":false,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowDropBox-true",
				Input: "{\"AllowDropBox\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":true,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowBox-null",
				Input: "{\"AllowBox\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowBox-false",
				Input: "{\"AllowBox\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":false,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowBox-true",
				Input: "{\"AllowBox\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":true,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowGoogleDrive-null",
				Input: "{\"AllowGoogleDrive\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowGoogleDrive-false",
				Input: "{\"AllowGoogleDrive\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":false,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowGoogleDrive-true",
				Input: "{\"AllowGoogleDrive\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":true,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowShareFile-null",
				Input: "{\"AllowShareFile\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowShareFile-false",
				Input: "{\"AllowShareFile\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":false,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowShareFile-true",
				Input: "{\"AllowShareFile\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":true,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowEgnyte-null",
				Input: "{\"AllowEgnyte\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowEgnyte-false",
				Input: "{\"AllowEgnyte\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":false,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowEgnyte-true",
				Input: "{\"AllowEgnyte\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":true,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowOrganizationTab-null",
				Input: "{\"AllowOrganizationTab\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowOrganizationTab-false",
				Input: "{\"AllowOrganizationTab\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":false,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowOrganizationTab-true",
				Input: "{\"AllowOrganizationTab\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":true,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowSkypeBusinessInterop-null",
				Input: "{\"AllowSkypeBusinessInterop\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowSkypeBusinessInterop-false",
				Input: "{\"AllowSkypeBusinessInterop\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":false,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowSkypeBusinessInterop-true",
				Input: "{\"AllowSkypeBusinessInterop\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":true,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ContentPin-null",
				Input: "{\"ContentPin\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ContentPin-empty",
				Input: "{\"ContentPin\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":\"\",\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ContentPin-future",
				Input: "{\"ContentPin\":\"Future-ContentPin\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":\"Future-ContentPin\",\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowResourceAccountSendMessage-null",
				Input: "{\"AllowResourceAccountSendMessage\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowResourceAccountSendMessage-false",
				Input: "{\"AllowResourceAccountSendMessage\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":false,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowResourceAccountSendMessage-true",
				Input: "{\"AllowResourceAccountSendMessage\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":true,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ResourceAccountContentAccess-null",
				Input: "{\"ResourceAccountContentAccess\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ResourceAccountContentAccess-empty",
				Input: "{\"ResourceAccountContentAccess\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":\"\",\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ResourceAccountContentAccess-future",
				Input: "{\"ResourceAccountContentAccess\":\"Future-ResourceAccountContentAccess\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":\"Future-ResourceAccountContentAccess\",\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowGuestUser-null",
				Input: "{\"AllowGuestUser\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowGuestUser-false",
				Input: "{\"AllowGuestUser\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":false,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowGuestUser-true",
				Input: "{\"AllowGuestUser\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":true,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowScopedPeopleSearchandAccess-null",
				Input: "{\"AllowScopedPeopleSearchandAccess\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowScopedPeopleSearchandAccess-false",
				Input: "{\"AllowScopedPeopleSearchandAccess\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":false,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowScopedPeopleSearchandAccess-true",
				Input: "{\"AllowScopedPeopleSearchandAccess\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":true,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowRoleBasedChatPermissions-null",
				Input: "{\"AllowRoleBasedChatPermissions\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowRoleBasedChatPermissions-false",
				Input: "{\"AllowRoleBasedChatPermissions\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":false,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "AllowRoleBasedChatPermissions-true",
				Input: "{\"AllowRoleBasedChatPermissions\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":true,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ExtendedWorkInfoInPeopleSearch-null",
				Input: "{\"ExtendedWorkInfoInPeopleSearch\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ExtendedWorkInfoInPeopleSearch-false",
				Input: "{\"ExtendedWorkInfoInPeopleSearch\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":false,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "ExtendedWorkInfoInPeopleSearch-true",
				Input: "{\"ExtendedWorkInfoInPeopleSearch\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":true,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "UseUnifiedDomain-null",
				Input: "{\"UseUnifiedDomain\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "UseUnifiedDomain-empty",
				Input: "{\"UseUnifiedDomain\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":\"\"}"),
				},
			},
			{
				Name:  "UseUnifiedDomain-future",
				Input: "{\"UseUnifiedDomain\":\"Future-UseUnifiedDomain\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":\"Future-UseUnifiedDomain\"}"),
				},
			},
			{
				Name:  "long-sender-list",
				Input: "{\"RestrictedSenderList\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":\"" + strings.Repeat("x", 2000) + "\",\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "opaque-sender-text",
				Input: "{\"RestrictedSenderList\":\"alpha@example.invalid; beta@example.invalid\",\"ContentPin\":\"0000\",\"ResourceAccountContentAccess\":\"FutureAccess\",\"UseUnifiedDomain\":\"FutureMode\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":\"0000\",\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":\"FutureAccess\",\"RestrictedSenderList\":\"alpha@example.invalid; beta@example.invalid\",\"UseUnifiedDomain\":\"FutureMode\"}"),
				},
			},
			{
				Name:  "case-duplicate",
				Input: "{\"RestrictedSenderList\":\"first\",\"restrictedsenderlist\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":\"second\",\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "duplicate-null",
				Input: "{\"AllowEmailIntoChannel\":true,\"AllowEmailIntoChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowBox\":null,\"AllowDropBox\":null,\"AllowEgnyte\":null,\"AllowEmailIntoChannel\":null,\"AllowGoogleDrive\":null,\"AllowGuestUser\":null,\"AllowOrganizationTab\":null,\"AllowResourceAccountSendMessage\":null,\"AllowRoleBasedChatPermissions\":null,\"AllowScopedPeopleSearchandAccess\":null,\"AllowShareFile\":null,\"AllowSkypeBusinessInterop\":null,\"ContentPin\":null,\"ExtendedWorkInfoInPeopleSearch\":null,\"Identity\":null,\"ResourceAccountContentAccess\":null,\"RestrictedSenderList\":null,\"UseUnifiedDomain\":null}"),
				},
			},
			{
				Name:  "invalid-boolean-object",
				Input: "{\"AllowEmailIntoChannel\":{}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
			{
				Name:  "invalid-string-array",
				Input: "{\"RestrictedSenderList\":[]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
		},
	}

	return fixture.Models
}
func exactJSON(t *testing.T, raw []byte) any {
	t.Helper()
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestUnitClientConfigurationKnownProjectionsMatchTypedModels(t *testing.T) {
	t.Parallel()
	for _, m := range modelCases(t) {
		t.Run(m.Name, func(t *testing.T) {

			var value resource.Configuration
			var err error
			if m.Input != "constructor" {
				err = json.Unmarshal([]byte(m.Input), &value)
			}
			if m.Observation.Failure != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source-invalid model accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("source-valid typed model rejected")
			}
			type known resource.Configuration
			raw, err := json.Marshal(known(value))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exactJSON(t, raw), exactJSON(t, m.Observation.Value)) {
				t.Fatal("known model projection differs")
			}
		})
	}
}
func TestUnitClientConfigurationPreservesOwnedRawSnapshot(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","AllowEmailIntoChannel":false,"ContentPin":"0000","Future":{"n":9007199254740993},"RestrictedSenderList":"first","restrictedsenderlist":null}`)
	saved := append([]byte(nil), raw...)
	var value resource.Configuration
	if json.Unmarshal(raw, &value) != nil || value.AllowEmailIntoChannel == nil || *value.AllowEmailIntoChannel || value.ContentPin == nil || *value.ContentPin != "0000" || value.RestrictedSenderList != nil {
		t.Fatal("nullable/opaque projection differs")
	}
	clear(raw)
	value.RawFields["Future"][0] = 'x'
	got, err := json.Marshal(value)
	if err != nil || !bytes.Equal(got, saved) {
		t.Fatal("snapshot lost bytes or ownership")
	}
}

func TestUnitClientConfigurationRejectsUnsupportedSourceShapes(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"AllowEmailIntoChannel":"yes"}`, `{"RestrictedSenderList":[]}`, `{"ContentPin":{}}`, `null`, `[]`} {
		var value resource.Configuration
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("source-rejected JSON shape accepted")
		}
	}
}

func TestUnitClientConfigurationRequestsRetainTypedModelSettings(t *testing.T) {
	t.Parallel()

	for index, m := range modelCases(t) {
		if m.Observation.Failure != nil || m.Input == "constructor" || m.Input == "{}" {
			continue
		}
		var input map[string]json.RawMessage
		var known map[string]json.RawMessage
		if json.Unmarshal([]byte(m.Input), &input) != nil || json.Unmarshal(m.Observation.Value, &known) != nil {
			t.Fatal("invalid model input/projection")
		}
		bound := map[string]json.RawMessage{}
		for key := range input {
			for canonical, value := range known {
				if strings.EqualFold(key, canonical) {
					bound[canonical] = value
					break
				}
			}
		}
		raw, err := json.Marshal(bound)
		if err != nil {
			t.Fatal(err)
		}
		var settings resource.Settings
		if json.Unmarshal(raw, &settings) != nil {
			t.Fatal("accepted final settings not representable")
		}
		bound["Identity"] = json.RawMessage(`"Global"`)
		body, err := json.Marshal(bound)
		if err != nil {
			t.Fatal(err)
		}
		run(t, contracttest.Case{Name: fmt.Sprintf("typed-model-request-%03d", index), Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", item, setCommand, body, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsClientConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings, Controls: policy.Controls{}})
			return contracttest.Observation{Response: r, Err: e}
		}, Want: contracttest.Expected{APIError: &client.
			APIError{Backend: client.TeamsAdmin,
			StatusCode: 400}},
		})
	}
}
