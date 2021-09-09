package chaos

// Duck can Quak at K8s resources.
type Duck interface {
	// Quak at K8s resources destroying them.
	Quak()

	takenCareOf(*stateCtx, synchronized) Duck
}

