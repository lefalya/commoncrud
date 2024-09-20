package types

type PaginationError struct {
	Err     error
	Details string
	Message string
}
