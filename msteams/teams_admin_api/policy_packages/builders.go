package policypackages

import "strings"

// ParsePolicyList converts the custom wrapper's Type/Name strings locally. Only
// the first comma, dot, colon, semicolon, space or tab is consumed; each resulting
// piece is then trimmed. Empty pieces and duplicates are retained. No HTTP runs.
// Group rankings use the native typed PolicyRanking API to avoid implicit shell
// numeric coercion; this helper does not infer ranks from policy names.
func ParsePolicyList(entries []string) ([]PolicyPair, error) {
	if len(entries) == 0 {
		return nil, ErrInput
	}
	result := make([]PolicyPair, 0, len(entries))
	for _, entry := range entries {
		index := strings.IndexAny(entry, ",.:; \t")
		if index < 0 {
			return nil, ErrInput
		}
		kind, name := strings.TrimSpace(entry[:index]), strings.TrimSpace(entry[index+1:])
		result = append(result, PolicyPair{PolicyType: &kind, PolicyName: &name})
	}
	return result, nil
}
