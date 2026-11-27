package tenantconfiguration

import "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"

// validate mirrors nullable UInt32 minimum-one model setters. The uint32 type
// bounds the upper end; null is distinct from the rejected explicit zero.
func validate(value Settings) error {
	for _, field := range []policy.Field[uint32]{value.MaxAllowedDomains, value.MaxBlockedDomains} {
		if field.IsSet() && !field.IsNull() {
			v, _ := field.Value()
			if v < 1 {
				return ErrInput
			}
		}
	}
	return nil
}
