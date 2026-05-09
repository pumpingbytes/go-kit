package dberror

// Classifier classifies storage-layer errors into common categories that
// repositories and service adapters can map into domain or application errors.
type Classifier interface {
	IsNoRowsError(err error) bool
	IsForeignKeyViolationError(err error) bool
	IsUniqueViolationError(err error) bool
}
