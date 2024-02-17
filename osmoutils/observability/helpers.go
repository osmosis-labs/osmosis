package observability

func FormatMetricName(moduleName, extension string) string {
	return moduleName + "_" + extension
}
