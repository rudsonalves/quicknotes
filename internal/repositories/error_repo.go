package repositories

import "log/slog"

type RepositoriesError struct {
	error
}

func newRepositoryError(err error) error {
	return RepositoriesError{error: err}
}

func fail(err error) error {
	slog.Error(err.Error())
	return newRepositoryError(err)
}
