// Copyright 2026 Canonical Ltd
// SPDX-License-Identifier: Apache-2.0

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
