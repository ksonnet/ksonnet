package slice

import "github.com/clipperhouse/typewriter"

var aggregateT = &typewriter.Template{
	Name: "Aggregate",
	Text: `
// Aggregate{{.TypeParameter.LongName}} iterates over {{.SliceName}}, operating on each element while maintaining ‘state’. See: http://clipperhouse.github.io/gen/#Aggregate
func (rcv {{.SliceName}}) Aggregate{{.TypeParameter.LongName}}(fn func({{.TypeParameter}}, {{.Type}}) {{.TypeParameter}}) (result {{.TypeParameter}}) {
	for _, v := range rcv {
		result = fn(result, v)
	}
	return
}
`,
	TypeParameterConstraints: []typewriter.Constraint{
		// exactly one type parameter is required, but no constraints on that type
		{},
	},
}
