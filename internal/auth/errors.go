package auth

// AuthError is a credential problem: missing profile, undecryptable token,
// or an unusable keyring backend.
type AuthError struct {
	Reason string
}

func (e *AuthError) Error() string { return e.Reason }

// ExitCode returns the process exit code for auth failures.
func (e *AuthError) ExitCode() int { return 3 }
