package teamschannelspolicy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_channels_policy"
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
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "all-settings",
				Input: "{\"Description\":\"Future-Description\",\"AllowOrgWideTeamCreation\":false,\"EnablePrivateTeamDiscovery\":false,\"AllowPrivateChannelCreation\":false,\"AllowSharedChannelCreation\":false,\"AllowChannelSharingToExternalUser\":false,\"AllowUserToParticipateInExternalSharedChannel\":false,\"ThreadedChannelCreation\":\"Future-ThreadedChannelCreation\",\"AllowCreateChannel\":\"Future-AllowCreateChannel\",\"AllowCreateClassicChannel\":\"Future-AllowCreateClassicChannel\",\"AllowCreatePrivateChannel\":\"Future-AllowCreatePrivateChannel\",\"AllowCreateSharedChannel\":\"Future-AllowCreateSharedChannel\",\"CreateSharedChannelsByDefault\":\"Future-CreateSharedChannelsByDefault\",\"AllowUsersFromOutsideTeam\":\"Future-AllowUsersFromOutsideTeam\",\"AllowGuestsFromOutsideTeam\":\"Future-AllowGuestsFromOutsideTeam\",\"AllowUsersFromOutsideTeamInPrivateChannel\":\"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\"AllowSharingWithTeamInOrg\":\"Future-AllowSharingWithTeamInOrg\",\"AllowSharingPrivateChannelWithTeamInOrg\":\"Future-AllowSharingPrivateChannelWithTeamInOrg\",\"DefaultChannelTypeOnCreation\":\"Future-DefaultChannelTypeOnCreation\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":false,\"AllowCreateChannel\":\"Future-AllowCreateChannel\",\"AllowCreateClassicChannel\":\"Future-AllowCreateClassicChannel\",\"AllowCreatePrivateChannel\":\"Future-AllowCreatePrivateChannel\",\"AllowCreateSharedChannel\":\"Future-AllowCreateSharedChannel\",\"AllowGuestsFromOutsideTeam\":\"Future-AllowGuestsFromOutsideTeam\",\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\"AllowOrgWideTeamCreation\":false,\"AllowPrivateChannelCreation\":false,\"AllowSharedChannelCreation\":false,\"AllowSharingPrivateChannelWithTeamInOrg\":\"Future-AllowSharingPrivateChannelWithTeamInOrg\",\"AllowSharingWithTeamInOrg\":\"Future-AllowSharingWithTeamInOrg\",\"AllowUserToParticipateInExternalSharedChannel\":false,\"AllowUsersFromOutsideTeam\":\"Future-AllowUsersFromOutsideTeam\",\"AllowUsersFromOutsideTeamInPrivateChannel\":\"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\"CreateSharedChannelsByDefault\":\"Future-CreateSharedChannelsByDefault\",\"DefaultChannelTypeOnCreation\":\"Future-DefaultChannelTypeOnCreation\",\"Description\":\"Future-Description\",\"EnablePrivateTeamDiscovery\":false,\"Identity\":null,\"ThreadedChannelCreation\":\"Future-ThreadedChannelCreation\"}"),
				},
			},
			{
				Name:  "Description-null",
				Input: "{\"Description\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "Description-empty",
				Input: "{\"Description\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":\"\",\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "Description-future",
				Input: "{\"Description\":\"Future-Description\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":\"Future-Description\",\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowOrgWideTeamCreation-null",
				Input: "{\"AllowOrgWideTeamCreation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowOrgWideTeamCreation-false",
				Input: "{\"AllowOrgWideTeamCreation\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":false,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowOrgWideTeamCreation-true",
				Input: "{\"AllowOrgWideTeamCreation\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":true,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "EnablePrivateTeamDiscovery-null",
				Input: "{\"EnablePrivateTeamDiscovery\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "EnablePrivateTeamDiscovery-false",
				Input: "{\"EnablePrivateTeamDiscovery\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":false,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "EnablePrivateTeamDiscovery-true",
				Input: "{\"EnablePrivateTeamDiscovery\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":true,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowPrivateChannelCreation-null",
				Input: "{\"AllowPrivateChannelCreation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowPrivateChannelCreation-false",
				Input: "{\"AllowPrivateChannelCreation\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":false,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowPrivateChannelCreation-true",
				Input: "{\"AllowPrivateChannelCreation\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":true,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharedChannelCreation-null",
				Input: "{\"AllowSharedChannelCreation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharedChannelCreation-false",
				Input: "{\"AllowSharedChannelCreation\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":false,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharedChannelCreation-true",
				Input: "{\"AllowSharedChannelCreation\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":true,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowChannelSharingToExternalUser-null",
				Input: "{\"AllowChannelSharingToExternalUser\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowChannelSharingToExternalUser-false",
				Input: "{\"AllowChannelSharingToExternalUser\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":false,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowChannelSharingToExternalUser-true",
				Input: "{\"AllowChannelSharingToExternalUser\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":true,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUserToParticipateInExternalSharedChannel-null",
				Input: "{\"AllowUserToParticipateInExternalSharedChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUserToParticipateInExternalSharedChannel-false",
				Input: "{\"AllowUserToParticipateInExternalSharedChannel\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":false,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUserToParticipateInExternalSharedChannel-true",
				Input: "{\"AllowUserToParticipateInExternalSharedChannel\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":true,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "ThreadedChannelCreation-null",
				Input: "{\"ThreadedChannelCreation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "ThreadedChannelCreation-empty",
				Input: "{\"ThreadedChannelCreation\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":\"\"}"),
				},
			},
			{
				Name:  "ThreadedChannelCreation-future",
				Input: "{\"ThreadedChannelCreation\":\"Future-ThreadedChannelCreation\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":\"Future-ThreadedChannelCreation\"}"),
				},
			},
			{
				Name:  "AllowCreateChannel-null",
				Input: "{\"AllowCreateChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateChannel-empty",
				Input: "{\"AllowCreateChannel\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":\"\",\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateChannel-future",
				Input: "{\"AllowCreateChannel\":\"Future-AllowCreateChannel\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":\"Future-AllowCreateChannel\",\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateClassicChannel-null",
				Input: "{\"AllowCreateClassicChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateClassicChannel-empty",
				Input: "{\"AllowCreateClassicChannel\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":\"\",\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateClassicChannel-future",
				Input: "{\"AllowCreateClassicChannel\":\"Future-AllowCreateClassicChannel\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":\"Future-AllowCreateClassicChannel\",\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreatePrivateChannel-null",
				Input: "{\"AllowCreatePrivateChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreatePrivateChannel-empty",
				Input: "{\"AllowCreatePrivateChannel\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":\"\",\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreatePrivateChannel-future",
				Input: "{\"AllowCreatePrivateChannel\":\"Future-AllowCreatePrivateChannel\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":\"Future-AllowCreatePrivateChannel\",\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateSharedChannel-null",
				Input: "{\"AllowCreateSharedChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateSharedChannel-empty",
				Input: "{\"AllowCreateSharedChannel\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":\"\",\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowCreateSharedChannel-future",
				Input: "{\"AllowCreateSharedChannel\":\"Future-AllowCreateSharedChannel\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":\"Future-AllowCreateSharedChannel\",\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "CreateSharedChannelsByDefault-null",
				Input: "{\"CreateSharedChannelsByDefault\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "CreateSharedChannelsByDefault-empty",
				Input: "{\"CreateSharedChannelsByDefault\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":\"\",\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "CreateSharedChannelsByDefault-future",
				Input: "{\"CreateSharedChannelsByDefault\":\"Future-CreateSharedChannelsByDefault\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":\"Future-CreateSharedChannelsByDefault\",\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUsersFromOutsideTeam-null",
				Input: "{\"AllowUsersFromOutsideTeam\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUsersFromOutsideTeam-empty",
				Input: "{\"AllowUsersFromOutsideTeam\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":\"\",\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUsersFromOutsideTeam-future",
				Input: "{\"AllowUsersFromOutsideTeam\":\"Future-AllowUsersFromOutsideTeam\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":\"Future-AllowUsersFromOutsideTeam\",\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowGuestsFromOutsideTeam-null",
				Input: "{\"AllowGuestsFromOutsideTeam\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowGuestsFromOutsideTeam-empty",
				Input: "{\"AllowGuestsFromOutsideTeam\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":\"\",\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowGuestsFromOutsideTeam-future",
				Input: "{\"AllowGuestsFromOutsideTeam\":\"Future-AllowGuestsFromOutsideTeam\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":\"Future-AllowGuestsFromOutsideTeam\",\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUsersFromOutsideTeamInPrivateChannel-null",
				Input: "{\"AllowUsersFromOutsideTeamInPrivateChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUsersFromOutsideTeamInPrivateChannel-empty",
				Input: "{\"AllowUsersFromOutsideTeamInPrivateChannel\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":\"\",\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowUsersFromOutsideTeamInPrivateChannel-future",
				Input: "{\"AllowUsersFromOutsideTeamInPrivateChannel\":\"Future-AllowUsersFromOutsideTeamInPrivateChannel\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":\"Future-AllowUsersFromOutsideTeamInPrivateChannel\",\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowGuestsFromOutsideTeamInPrivateChannel-null",
				Input: "{\"AllowGuestsFromOutsideTeamInPrivateChannel\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowGuestsFromOutsideTeamInPrivateChannel-empty",
				Input: "{\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"\",\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowGuestsFromOutsideTeamInPrivateChannel-future",
				Input: "{\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"Future-AllowGuestsFromOutsideTeamInPrivateChannel\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":\"Future-AllowGuestsFromOutsideTeamInPrivateChannel\",\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharingWithTeamInOrg-null",
				Input: "{\"AllowSharingWithTeamInOrg\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharingWithTeamInOrg-empty",
				Input: "{\"AllowSharingWithTeamInOrg\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":\"\",\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharingWithTeamInOrg-future",
				Input: "{\"AllowSharingWithTeamInOrg\":\"Future-AllowSharingWithTeamInOrg\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":\"Future-AllowSharingWithTeamInOrg\",\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharingPrivateChannelWithTeamInOrg-null",
				Input: "{\"AllowSharingPrivateChannelWithTeamInOrg\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharingPrivateChannelWithTeamInOrg-empty",
				Input: "{\"AllowSharingPrivateChannelWithTeamInOrg\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":\"\",\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "AllowSharingPrivateChannelWithTeamInOrg-future",
				Input: "{\"AllowSharingPrivateChannelWithTeamInOrg\":\"Future-AllowSharingPrivateChannelWithTeamInOrg\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":\"Future-AllowSharingPrivateChannelWithTeamInOrg\",\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "DefaultChannelTypeOnCreation-null",
				Input: "{\"DefaultChannelTypeOnCreation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "DefaultChannelTypeOnCreation-empty",
				Input: "{\"DefaultChannelTypeOnCreation\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":\"\",\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "DefaultChannelTypeOnCreation-future",
				Input: "{\"DefaultChannelTypeOnCreation\":\"Future-DefaultChannelTypeOnCreation\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":\"Future-DefaultChannelTypeOnCreation\",\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "long-description",
				Input: "{\"Description\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "case-duplicate",
				Input: "{\"Description\":\"first\",\"description\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":\"second\",\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "duplicate-null",
				Input: "{\"AllowPrivateChannelCreation\":true,\"AllowPrivateChannelCreation\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":null,\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":null,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "distinct-legacy-boolean-and-string-mode",
				Input: "{\"AllowPrivateChannelCreation\":false,\"AllowCreatePrivateChannel\":\"FutureMode\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowChannelSharingToExternalUser\":null,\"AllowCreateChannel\":null,\"AllowCreateClassicChannel\":null,\"AllowCreatePrivateChannel\":\"FutureMode\",\"AllowCreateSharedChannel\":null,\"AllowGuestsFromOutsideTeam\":null,\"AllowGuestsFromOutsideTeamInPrivateChannel\":null,\"AllowOrgWideTeamCreation\":null,\"AllowPrivateChannelCreation\":false,\"AllowSharedChannelCreation\":null,\"AllowSharingPrivateChannelWithTeamInOrg\":null,\"AllowSharingWithTeamInOrg\":null,\"AllowUserToParticipateInExternalSharedChannel\":null,\"AllowUsersFromOutsideTeam\":null,\"AllowUsersFromOutsideTeamInPrivateChannel\":null,\"CreateSharedChannelsByDefault\":null,\"DefaultChannelTypeOnCreation\":null,\"Description\":null,\"EnablePrivateTeamDiscovery\":null,\"Identity\":null,\"ThreadedChannelCreation\":null}"),
				},
			},
			{
				Name:  "invalid-boolean-object",
				Input: "{\"AllowPrivateChannelCreation\":{}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
			{
				Name:  "invalid-string-object",
				Input: "{\"AllowCreatePrivateChannel\":{}}",
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
func TestUnitChannelsPolicyKnownProjectionsMatchTypedModels(t *testing.T) {
	t.Parallel()
	for _, m := range modelCases(t) {
		t.Run(m.Name, func(t *testing.T) {

			var value resource.Policy
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
			type known resource.Policy
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
func TestUnitChannelsPolicyPreservesOwnedRawSnapshot(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","AllowPrivateChannelCreation":false,"AllowCreatePrivateChannel":"FutureMode","Future":{"n":9007199254740993},"Description":"first","description":null}`)
	saved := append([]byte(nil), raw...)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.AllowPrivateChannelCreation == nil || *value.AllowPrivateChannelCreation || value.AllowCreatePrivateChannel == nil || *value.AllowCreatePrivateChannel != "FutureMode" || value.Description != nil {
		t.Fatal("distinct/nullable projection differs")
	}
	clear(raw)
	value.RawFields["Future"][0] = 'x'
	got, err := json.Marshal(value)
	if err != nil || !bytes.Equal(got, saved) {
		t.Fatal("snapshot lost bytes or ownership")
	}
}

func TestUnitChannelsPolicyRequiresNativeJSONTypes(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"AllowPrivateChannelCreation":"false"}`, `{"AllowCreatePrivateChannel":false}`, `{"Description":{}}`, `null`, `[]`} {
		var value resource.Policy
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("non-native JSON type accepted")
		}
	}
}

func TestUnitChannelsPolicyRequestsRetainTypedModelSettings(t *testing.T) {
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
		run(t, testutil.Case{Name: fmt.Sprintf("typed-model-request-%03d", index), Steps: []testutil.Exchange{compiledWire("PATCH", item, setCommand, body, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsChannelsPolicies.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings, Controls: policy.Controls{}})
			return testutil.Observation{Response: r, Err: e}
		}, Want: testutil.Expected{APIError: &client.
			APIError{Backend: client.
			TeamsAdmin, StatusCode: 400},
		},
		})
	}
}
