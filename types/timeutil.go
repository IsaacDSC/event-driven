package types

import "time"

type TimeProvider interface {
	Now() time.Time
}
