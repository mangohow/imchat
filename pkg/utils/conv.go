package utils

func ToInterfaceSlice[T any](ss []T) []interface{} {
	res := make([]interface{}, 0, len(ss))
	for i := range ss {
		res = append(res, ss[i])
	}

	return res
}
