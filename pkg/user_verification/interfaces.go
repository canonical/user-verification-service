package user_verification

import (
	"context"
)

type ServiceInterface interface {
	IsEmployee(context.Context, string) (bool, error)
}
