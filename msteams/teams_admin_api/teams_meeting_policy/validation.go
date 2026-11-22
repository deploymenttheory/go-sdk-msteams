package teamsmeetingpolicy

import (
	"regexp"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// The source setters accept GUID text or null and a nullable bit rate >= 1.
// .NET's $ anchor also accepts one final newline; preserve that binding behavior.
var guidSetting = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}(?:\n)?$`)

func validateSettings(s Settings) error {
	for _, field := range []policy.Field[string]{s.RecordingAndTranscriptionCustomMessageIdentifier, s.SyntheticMediaDetectionAppId, s.PreMeetingConsentContentIdentifier} {
		if value, ok := field.Value(); ok && !guidSetting.MatchString(value) {
			return ErrInput
		}
	}
	if value, ok := s.MediaBitRateKb.Value(); ok && value < 1 {
		return ErrInput
	}
	return nil
}
