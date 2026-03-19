package errs
import "errors"
// implement me
var (
ErrNotFound = errors.New("not found")
ErrConflict = errors.New("conflict")
)
