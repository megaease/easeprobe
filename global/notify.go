package global

// NotifySettings is the global notification setting
type NotifySettings struct {
	TimeFormat string
	Retry      Retry
}

// NormalizeRetry return a normalized retry value
func (n *NotifySettings) NormalizeRetry(retry Retry) Retry {
	retry.Interval = normalizeTimeDuration(n.Retry.Interval, retry.Interval, 0, DefaultRetryInterval)
	retry.Times = normalizeInteger(n.Retry.Times, retry.Times, 0, DefaultRetryTimes)
	return retry
}
