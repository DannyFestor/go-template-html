package realtime

import "strings"

const (
	segmentSeparator = "."
	paramOpen        = "{"
	paramClose       = "}"
)

type channelRule struct {
	segments  []string
	private   bool
	authorize Authorize
}

func newChannelRule(pattern string, private bool, authorize Authorize) channelRule {
	return channelRule{segments: strings.Split(pattern, segmentSeparator), private: private, authorize: authorize}
}

func (r channelRule) match(channel string) (map[string]string, bool) {
	segments := strings.Split(channel, segmentSeparator)
	if len(segments) != len(r.segments) {
		return nil, false
	}

	params := map[string]string{}

	for i, patternSegment := range r.segments {
		name, isParam := paramName(patternSegment)

		switch {
		case segments[i] == "":
			return nil, false
		case isParam:
			params[name] = segments[i]
		case segments[i] != patternSegment:
			return nil, false
		}
	}

	return params, true
}

func paramName(segment string) (string, bool) {
	if !strings.HasPrefix(segment, paramOpen) || !strings.HasSuffix(segment, paramClose) {
		return "", false
	}

	return strings.TrimSuffix(strings.TrimPrefix(segment, paramOpen), paramClose), true
}
