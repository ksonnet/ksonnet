package slice

import "github.com/clipperhouse/typewriter"

var sortInterface = &typewriter.Template{
	Name: "sortInterface",
	Text: `
func (rcv {{.SliceName}}) Len() int {
	return len(rcv)
}
func (rcv {{.SliceName}}) Less(i, j int) bool {
	return rcv[i] < rcv[j]
}
func (rcv {{.SliceName}}) Swap(i, j int) {
	rcv[i], rcv[j] = rcv[j], rcv[i]
}
`}
