// Copyright 2026 Canonical Ltd
// SPDX-License-Identifier: AGPL-3.0-only

package salesforce

import (
	"context"
)

type SalesforceAPI interface {
	IsEmployee(context.Context, string) (bool, error)
}

type SalesforceClientAPI interface {
	Query(string, any) error
}
