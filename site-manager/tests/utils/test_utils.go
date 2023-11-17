package test_utils

// ToSliceOfInterfaces converts array of objects to array of interfaces
func ToSliceOfInterfaces[T any](objects []T) []interface{} {
	result := make([]interface{}, len(objects))
	for i, obj := range objects {
		result[i] = obj
	}
	return result
}
