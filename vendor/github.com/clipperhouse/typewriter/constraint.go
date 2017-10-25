package typewriter

import "fmt"

// Constraint describes type requirements.
type Constraint struct {
	// A numeric type is one that supports arithmetic operations.
	Numeric bool
	// A comparable type is one that supports the == operator. Map keys must be comparable, for example.
	Comparable bool
	// An ordered type is one where greater-than and less-than are supported
	Ordered bool
}

func (c Constraint) TryType(t Type) error {
	if c.Comparable && !t.comparable {
		return fmt.Errorf("%s must be comparable (i.e. support == and !=)", t)
	}

	if c.Numeric && !t.numeric {
		return fmt.Errorf("%s must be numeric", t)
	}

	if c.Ordered && !t.ordered {
		return fmt.Errorf("%s must be ordered (i.e. support > and <)", t)
	}

	return nil
}
