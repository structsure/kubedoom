package results

import "strings"

type Results []string

func (r Results) Only(matching string) Results {
	return r.filterResults(
		func(inThisString string) bool {
			return strings.Contains(inThisString, matching)
		})
}
func (r Results) Not(matching string) Results {
	return r.filterResults(
		func(inThisString string) bool {
			return !strings.Contains(inThisString, matching)
		})
}
func (r Results) filterResults(filter func(string) bool) Results {
	filtered := []string{}
	for iSlice, vSlice := range r {
		if filter(vSlice) {
			filtered = append(filtered, r[iSlice])
		}
	}
	return filtered
}
