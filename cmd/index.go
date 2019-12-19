package cmd

// Range is a text serialisable width.
type Range struct{ Min, Max int64 }

// Operation is a text serialisable lib.Operation.
type Operation struct {
	Name  string
	Range *Range
}
