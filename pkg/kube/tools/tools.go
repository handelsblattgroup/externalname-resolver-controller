package tools

import "strconv"

func FormatInt32(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

func LabelExists(labels map[string]string, label string) bool {
	return entryExists(labels, label)
}

func AnnotationExists(annotations map[string]string, annotation string) bool {
	return entryExists(annotations, annotation)
}

func entryExists(entries map[string]string, key string) bool {
	_, exists := entries[key]

	return exists
}

func AnnotationExistsAndIsEqual(annotations map[string]string, annotation, value string) bool {
	return entryExistsAndIsEqual(annotations, annotation, value)
}

func entryExistsAndIsEqual(entries map[string]string, key, value string) bool {
	if !entryExists(entries, key) {
		return false
	}

	return entries[key] == value
}
