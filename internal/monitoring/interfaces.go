package monitoring

type MonitorInterface interface {
	GetService() string
	SetDirectoryApiResponseTimeMetric(map[string]string, float64) error
	SetResponseTimeMetric(map[string]string, float64) error
	SetDependencyAvailability(map[string]string, float64) error
}
