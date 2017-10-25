package slice

import "github.com/clipperhouse/typewriter"

var shuffle = &typewriter.Template{
	Name: "Shuffle",
	Text: `
// Shuffle returns a shuffled copy of {{.SliceName}}, using a version of the Fisher-Yates shuffle. See: http://clipperhouse.github.io/gen/#Shuffle
func (rcv {{.SliceName}}) Shuffle() {{.SliceName}} {
    numItems := len(rcv)
    result := make({{.SliceName}}, numItems)
    copy(result, rcv)
    for i := 0; i < numItems; i++ {
        r := i + rand.Intn(numItems-i)
        result[r], result[i] = result[i], result[r]
    }
    return result
}
`}
