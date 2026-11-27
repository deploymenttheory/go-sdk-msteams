package teamscustombannertext_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/CustomBannerText"
const relative = "01234567-89ab-cdef-0123-456789abcdef"
const item = collection + "/configuration/Global/" + relative

const newCommand = "New-CsTeamsCustomBannerText"
const setCommand = "Set-CsTeamsCustomBannerText"
const deleteCommand = "Remove-CsTeamsCustomBannerText"

type inputVectors struct {
	Name, Operation, Selector, Identity, ID, Input, Response string
	Status                                                   int
	ForceMode                                                int    `json:"force_mode"`
	ReadID                                                   string `json:"read_model_id"`
	Observed                                                 struct {
		Identity string `json:"normalized_identity"`
		Request  struct {
			Method, Url, Body, CmdletName string
			FailureType                   *string
		} `json:"request"`
	} `json:"observed"`
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	fixture := struct{ Cases []inputVectors }{
		Cases: []inputVectors{
			{
				Name:      "read-identity",
				Operation: "read",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						CmdletName: "Get-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "read-relative",
				Operation: "read",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						CmdletName: "Get-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "read-extra-slash",
				Operation: "read",
				Selector:  "identity",
				Identity:  "scope /01234567-89ab-cdef-0123-456789abcdef / Child",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "scope /01234567-89ab-cdef-0123-456789abcdef / Child",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/scope%20/01234567-89ab-cdef-0123-456789abcdef%20%2F%20Child",
						CmdletName: "Get-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "list",
				Operation: "list",
				Selector:  "default",
				Identity:  "",
				ID:        "",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						CmdletName: "Get-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "create-identity",
				Operation: "create",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"fixture\",\"Description\":\"\"}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "POST",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						Body:       "{\n  \"Identity\": \"Global/01234567-89ab-cdef-0123-456789abcdef\",\n  \"Text\": \"fixture\",\n  \"Description\": \"\",\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "New-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "create-default-id",
				Operation: "create",
				Selector:  "default",
				Identity:  "",
				ID:        "",
				Input:     "{\"Text\":\"fixture\"}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/b735f90d-0946-4b90-8e45-56ab483a0b08",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "POST",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						Body:       "{\n  \"Text\": \"fixture\",\n  \"Identity\": \"Global/b735f90d-0946-4b90-8e45-56ab483a0b08\"\n}",
						CmdletName: "New-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "create-relative",
				Operation: "create",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"fixture\",\"Description\":\"\"}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "POST",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						Body:       "{\n  \"Id\": null,\n  \"Text\": \"fixture\",\n  \"Description\": \"\",\n  \"Identity\": \"Global/01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "New-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "create-identity-force",
				Operation: "create",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"fixture\"}",

				Status:    400,
				ForceMode: 0,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "POST",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						Body:       "{\n  \"Identity\": \"Global/01234567-89ab-cdef-0123-456789abcdef\",\n  \"Text\": \"fixture\",\n  \"Force\": \"False\",\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "New-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "create-relative-force",
				Operation: "create",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"fixture\"}",

				Status:    400,
				ForceMode: 1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "POST",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						Body:       "{\n  \"Id\": null,\n  \"Text\": \"fixture\",\n  \"Force\": \"True\",\n  \"Identity\": \"Global/01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "New-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "create-relative-missing-text",
				Operation: "create",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						FailureType: new("System.ArgumentNullException"),
					},
				},
			},
			{
				Name:      "create-identity-description-null",
				Operation: "create",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"fixture\",\"Description\":null}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "POST",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						Body:       "{\n  \"Identity\": \"Global/01234567-89ab-cdef-0123-456789abcdef\",\n  \"Text\": \"fixture\",\n  \"Description\": null,\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "New-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "patch-identity",
				Operation: "patch",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"\",\"Description\":null}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "PATCH",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Text\": \"\",\n  \"Description\": null\n}",
						CmdletName: "Set-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "patch-relative",
				Operation: "patch",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":\"\",\"Description\":null}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "PATCH",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\",\n  \"Text\": \"\",\n  \"Description\": null\n}",
						CmdletName: "Set-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "patch-relative-null-id",
				Operation: "patch",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Description\":\"\"}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "PATCH",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Id\": null,\n  \"Description\": \"\"\n}",
						CmdletName: "Set-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "patch-relative-description-only",
				Operation: "patch",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Description\":\"\"}",

				Status:    400,
				ForceMode: 0,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "PATCH",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\",\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",
						CmdletName: "Set-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "patch-identity-force",
				Operation: "patch",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Description\":\"\"}",

				Status:    400,
				ForceMode: 1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "PATCH",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"Force\": \"True\"\n}",
						CmdletName: "Set-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "patch-text-null",
				Operation: "patch",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{\"Text\":null}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						FailureType: new("System.ArgumentNullException"),
					},
				},
			},
			{
				Name:      "delete-identity",
				Operation: "delete",
				Selector:  "identity",
				Identity:  "Global/01234567-89ab-cdef-0123-456789abcdef",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "DELETE",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Identity\": \"Global/01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "Remove-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "delete-relative",
				Operation: "delete",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "DELETE",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\"\n}",
						CmdletName: "Remove-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "delete-relative-force",
				Operation: "delete",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: 0,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "DELETE",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						Body:       "{\n  \"Id\": \"01234567-89ab-cdef-0123-456789abcdef\",\n  \"Force\": \"False\"\n}",
						CmdletName: "Remove-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "read-null",
				Operation: "read",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						CmdletName: "Get-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						CmdletName: "Get-CsTeamsCustomBannerText",
					},
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Selector:  "default",
				Identity:  "",
				ID:        "",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:      "GET",
						Url:         "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText",
						CmdletName:  "Get-CsTeamsCustomBannerText",
						FailureType: new("Newtonsoft.Json.JsonSerializationException"),
					},
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Selector:  "relative",
				Identity:  "",
				ID:        "01234567-89ab-cdef-0123-456789abcdef",
				Input:     "{}",

				Status:    400,
				ForceMode: -1,
				ReadID:    "01234567-89ab-cdef-0123-456789abcdef",
				Observed: struct {
					Identity string `json:"normalized_identity"`
					Request  struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					} `json:"request"`
				}{
					Identity: "Global/01234567-89ab-cdef-0123-456789abcdef",
					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:      "GET",
						Url:         "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/CustomBannerText/configuration/Global/01234567-89ab-cdef-0123-456789abcdef",
						CmdletName:  "Get-CsTeamsCustomBannerText",
						FailureType: new("Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRpException"),
					},
				},
			},
		},
	}

	return fixture.Cases
}

func wire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("banner contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}
