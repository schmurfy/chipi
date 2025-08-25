package shared

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

func GetPtr[T any](v T) *T {
	return &v
}
