package types

type PaginationError struct {
	Err     error
	Details string
	Message string
}

type SortingOption struct {
	Attribute string
	Direction string
	Index     int
}
