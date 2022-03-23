package global

// NotifySettings is the global notification setting
type NotifySettings struct {
	TimeFormat string
	Retry      Retry
}

// NormalizeRetry return a normalized retry value
func (n *NotifySettings) NormalizeRetry(retry Retry) Retry {

	// if the val is in valid, the assign the default value
	if retry.Interval <= 0 {
		retry.Interval = DefaultRetryInterval
		//if the global configuration is validated, assign the global
		if n.Retry.Interval > 0 {
			retry.Interval = n.Retry.Interval
		}
	}

	// if the val is in valid, the assign the default value
	if retry.Times <= 0 {
		retry.Times = DefaultRetryTimes
		if n.Retry.Times > 0 {
			retry.Times = n.Retry.Times
		}
	}

	return retry
}
