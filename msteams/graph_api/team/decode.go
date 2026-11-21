package team

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func decodeTeam(body []byte) (*Team, error) {
	var value Team
	if json.Unmarshal(body, &value) != nil {
		return nil, client.ErrDecode
	}
	return &value, nil
}

func decodePage(body []byte) (client.PageContent[Team], error) {
	return client.DecodeCollection[Team](body, "value", "@odata.nextLink")
}
