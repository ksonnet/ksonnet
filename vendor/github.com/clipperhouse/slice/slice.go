package slice

import "github.com/clipperhouse/typewriter"

var slice = &typewriter.Template{
	Name: "slice",
	Text: `// {{.SliceName}} is a slice of type {{.Type}}. Use it where you would use []{{.Type}}.
type {{.SliceName}} []{{.Type}}
`,
}
