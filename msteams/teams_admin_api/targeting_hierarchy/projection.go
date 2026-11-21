package targetinghierarchy

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func orderedKeys(raw json.RawMessage) ([]string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') {
		return nil, client.ErrDecode
	}
	var keys []string
	seen := map[string]bool{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, client.ErrDecode
		}
		name, ok := key.(string)
		if !ok {
			return nil, client.ErrDecode
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return nil, client.ErrDecode
		}
		if !seen[name] {
			seen[name] = true
			keys = append(keys, name)
		}
	}
	if _, err := decoder.Token(); err != nil {
		return nil, client.ErrDecode
	}
	return keys, nil
}

func lineEnding() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// Presentation follows the pinned CLR Result constructors, including argument
// iteration order and per-platform AppendLine behavior. Raw native data remains
// independent; malformed nested error text does not erase the readable status.
func projectRetail(value *RetailOperation) (*Status, error) {
	if value == nil {
		return nil, client.ErrDecode
	}
	result := &Status{Version: V1, ID: value.ID, Status: value.Status, RawStatus: value.Status, LastUpdatedAt: value.LastUpdatedAt, LastModifiedBy: value.LastModifiedBy, Filename: value.Filename, Retail: value}
	if value.ErrorDetails != nil {
		result.LastKnownStatus = value.ErrorDetails.LastKnownStatus
	}
	errors, err := value.DecodeErrors()
	if err != nil {
		return result, err
	}
	if len(errors) == 0 {
		return result, nil
	}
	var output strings.Builder
	_, _ = output.WriteString(strconv.Itoa(len(errors)) + " error(s) were found.\r\n")
	line := func(value string) { _, _ = output.WriteString(value + lineEnding()) }
	for _, item := range errors {
		line("Error: " + text(item.Code))
		line("Descriptions:")
		var rawArgs []json.RawMessage
		if item.Args == nil || json.Unmarshal(item.RawFields["args"], &rawArgs) != nil {
			return result, client.ErrDecode
		}
		for i, argument := range item.Args {
			keys, err := orderedKeys(rawArgs[i])
			if err != nil {
				return result, err
			}
			for _, key := range keys {
				if item.Description == nil {
					return result, client.ErrDecode
				}
				line(strings.ReplaceAll(*item.Description, "{"+key+"}", argument[key]))
			}
		}
	}
	message := output.String()
	result.ErrorMessage = &message
	return result, nil
}

func projectedStatus(raw string) string {
	trimmed := strings.TrimSpace(raw)
	values := map[string]int64{"NotStarted": 0, "Running": 1, "Succeeded": 2, "Failed": 3, "UnknownFutureValue": 4}
	value, known := values[trimmed]
	if !known {
		if strings.Contains(trimmed, ",") {
			// Enum.TryParse accepts comma-separated enum names even without a
			// Flags attribute. Preserve this presentation quirk without using
			// the projected string to infer native operation success.
			for _, part := range strings.Split(trimmed, ",") {
				member, ok := values[strings.TrimSpace(part)]
				if !ok {
					return raw
				}
				value |= member
			}
		} else {
			parsed, err := strconv.ParseInt(trimmed, 10, 32)
			if err != nil {
				return raw
			}
			value = parsed
		}
	}
	switch value {
	case 0:
		return "Starting"
	case 1:
		return "Ingesting"
	case 2:
		return "Successful"
	case 3:
		return "Failed"
	default:
		return "Unknown"
	}
}

func projectHierarchy(value *HierarchyOperation) (*Status, error) {
	if value == nil {
		return nil, client.ErrDecode
	}
	result := &Status{Version: V2, ID: value.ID, RawStatus: value.Status, LastUpdatedAt: value.LastActionDateTime, LastModifiedBy: value.LastModifiedBy, Filename: value.Filename, Hierarchy: value}
	if value.Status == nil {
		return result, client.ErrDecode
	}
	normalized := projectedStatus(*value.Status)
	result.Status = &normalized
	if value.Error == nil || len(value.Error.Details) == 0 {
		return result, nil
	}
	var output strings.Builder
	_, _ = output.WriteString(strconv.Itoa(len(value.Error.Details)) + " error(s) were found.\r\n")
	line := func(value string) { _, _ = output.WriteString(value + lineEnding()) }
	codes := []string{}
	messages := map[string][]string{}
	for _, item := range value.Error.Details {
		if item.Code == nil || item.Arguments == nil {
			return result, client.ErrDecode
		}
		keys, err := orderedKeys(item.RawFields["arguments"])
		if err != nil {
			return result, err
		}
		message := text(item.Message)
		for _, key := range keys {
			if item.Message == nil {
				return result, client.ErrDecode
			}
			message = strings.ReplaceAll(message, key, item.Arguments[key])
		}
		code := *item.Code
		if _, seen := messages[code]; !seen {
			codes = append(codes, code)
		}
		messages[code] = append(messages[code], message)
	}
	for _, code := range codes {
		line("Error: " + code)
		line("Descriptions:")
		for _, message := range messages[code] {
			line(message)
		}
	}
	message := output.String()
	result.ErrorMessage = &message
	return result, nil
}
