package serial

// Action is an action that can be run serially.
type Action func() error

// RunActions runs a set of actions serially. If an error occurs,
// it is returned and processing stops.
func RunActions(actions ...Action) error {
	for _, a := range actions {
		err := a()
		if err != nil {
			return err
		}
	}

	return nil
}
