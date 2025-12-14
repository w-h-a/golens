package wire

import "strings"

func extractAndCleanHeaders(headers map[string][]string) (map[string]string, map[string][]string) {
	attributes := map[string]string{}
	clean := map[string][]string{}

	for k, vv := range headers {
		lower := strings.ToLower(k)
		if strings.HasPrefix(lower, "golens-attribute-") {
			key := k[len("golens-attribute-"):]
			if len(vv) > 0 {
				attributes[key] = vv[0]
			}
		} else {
			clean[k] = append(clean[k], vv...)
		}
	}

	return attributes, clean
}
