package utils

import (
	"fmt"
	"sort"
	"strings"
)

// Add this function for deserialization:
func DeserializeLabels(s string) map[string]string {
	labels := make(map[string]string)
	if s == "" {
		return labels
	}
	for _, pair := range strings.Split(s, ";") {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

// ParseBucketKey splits "nodeID|name|labels" into parts and returns (nodeID, name, labels, error).
func ParseBucketKey(bucket string) (nodeID, name string, labels map[string]string, err error) {
	parts := strings.SplitN(bucket, "|", 3)
	if len(parts) != 3 {
		return "", "", nil, fmt.Errorf("bad bucket: %q", bucket)
	}
	nodeID, name, labelStr := parts[0], parts[1], parts[2]
	labels = DeserializeLabels(labelStr)
	return nodeID, name, labels, nil
}

// SerializeLabels is an *exported* version for use in tests and other packages.
func SerializeLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(labels[k])
		b.WriteByte(';')
	}
	return b.String()
}
