package directoryapi

import "context"

type DirectoryAPI interface {
	IsEmployee(context.Context, string) (bool, error)
}
