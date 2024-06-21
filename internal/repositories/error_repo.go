package repositories

type RepositoriesError struct {
	error
}

func newRepositoriesError(err error) error {
	return &RepositoriesError{error: err}
}
