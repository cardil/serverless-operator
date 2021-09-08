package test

import pkgupgrade "knative.dev/pkg/test/upgrade"

func Merge(slices ...[]pkgupgrade.BackgroundOperation) []pkgupgrade.BackgroundOperation {
	l := 0
	for _, slice := range slices {
		l += len(slice)
	}
	result := make([]pkgupgrade.BackgroundOperation, 0, l)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}
