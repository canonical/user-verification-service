package salesforce

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/user-verification-service/internal/logging"
	"github.com/canonical/user-verification-service/internal/monitoring"
	"github.com/canonical/user-verification-service/internal/tracing"
	"github.com/k-capehart/go-salesforce/v2"
)

var ErrInvalidTotalSize = fmt.Errorf("invalid total size")

const query = "SELECT fHCM2__Email__c, Employment_Record_Active__c FROM fHCM2__Team_Member__c WHERE fHCM2__Email__c = '%s' AND fHCM2__Has_Left__c = 'False'"

type Record struct {
	Employment_Record_Active__c bool
}

type Client struct {
	salesforceClient SalesforceClientAPI

	tracer  tracing.TracingInterface
	monitor monitoring.MonitorInterface
	logger  logging.LoggerInterface
}

func NewSalesforceClient(domain, consumerKey, consumerSecret string) (*salesforce.Salesforce, error) {
	return salesforce.Init(salesforce.Creds{
		Domain:         domain,
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
	})
}

func (c *Client) doAndMonitor(mail string) ([]Record, error) {
	startTime := time.Now()
	q := fmt.Sprintf(query, mail)
	r := []Record{}
	err := c.salesforceClient.Query(q, &r)

	tags := map[string]string{
		"user":  "*",
		"error": "nil",
	}
	if err != nil {
		c.logger.Errorf("failed to call salesforce: %s", err)
		tags["user"] = mail
		tags["error"] = err.Error()
	}
	c.monitor.SetSalesforceResponseTimeMetric(tags, time.Since(startTime).Seconds())

	return r, err
}

func (c *Client) IsEmployee(ctx context.Context, mail string) (bool, error) {
	recs, err := c.doAndMonitor(mail)
	if err != nil {
		return false, err
	}

	if len(recs) == 0 {
		c.logger.Errorf("Employee %s is inactive", mail)
		return false, nil
	}
	if len(recs) > 1 {
		c.logger.Errorf("Salesforce returned '%d' records, cannot parse result", len(recs))
		return false, ErrInvalidTotalSize
	}

	return recs[0].Employment_Record_Active__c, nil
}

func NewClient(
	domain, consumerKey, consumerSecret string,
	tracer tracing.TracingInterface,
	monitor monitoring.MonitorInterface,
	logger logging.LoggerInterface,
) *Client {
	var err error
	c := new(Client)

	c.salesforceClient, err = NewSalesforceClient(domain, consumerKey, consumerSecret)
	if err != nil {
		panic(fmt.Errorf("failed to initialize salesforce client: %v", err))
	}

	c.monitor = monitor
	c.tracer = tracer
	c.logger = logger

	return c
}
