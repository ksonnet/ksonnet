package slice

import "github.com/clipperhouse/typewriter"

var distinctBy = &typewriter.Template{
	Name: "DistinctBy",
	Text: `
// DistinctBy returns a new {{.SliceName}} whose elements are unique, where equality is defined by a passed func. See: http://clipperhouse.github.io/gen/#DistinctBy
func (rcv {{.SliceName}}) DistinctBy(equal func({{.Type}}, {{.Type}}) bool) (result {{.SliceName}}) {
Outer:
	for _, v := range rcv {
		for _, r := range result {
			if equal(v, r) {
				continue Outer
			}
		}
		result = append(result, v)
	}
	return result
}
`}
