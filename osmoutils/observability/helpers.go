package observability

// FormatMetricName helper to format a metric name given SDK module name and extension.
func FormatMetricName(moduleName, extension string) string {
	return moduleName + "_" + extension
}
