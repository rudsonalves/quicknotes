package repositories

type RepositoriesError struct {
	error
}

func newRepositoryError(err error) error {
	return RepositoriesError{error: err}
}
