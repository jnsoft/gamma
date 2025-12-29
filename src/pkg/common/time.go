package common

import "time"

func NowUnixUTC() int64 {
	return time.Now().UTC().Unix() // seconds since Jan 1, 1970 UTC
}

func NowUnixMilliUTC() int64 {
	return time.Now().UTC().UnixMilli() // milliseconds
}

func NowUnixNanoUTC() int64 {
	return time.Now().UTC().UnixNano() // nanoseconds
}

func TimeFromUnixUTC(sec int64) time.Time {
	return time.Unix(sec, 0).UTC()
}

// Convert milliseconds since epoch to time.Time (UTC)
func TimeFromUnixMilliUTC(ms int64) time.Time {
	return time.UnixMilli(ms).UTC()
}

// Convert nanoseconds since epoch to time.Time (UTC)
func TimeFromUnixNanoUTC(ns int64) time.Time {
	return time.Unix(0, ns).UTC()
}
