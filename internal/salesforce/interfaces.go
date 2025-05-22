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
