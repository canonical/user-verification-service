// Copyright 2026 Canonical Ltd
// SPDX-License-Identifier: Apache-2.0

package user_verification

import (
	"context"
)

type ServiceInterface interface {
	IsEmployee(context.Context, string) (bool, error)
}
