package core

import "time"

const (
	defaultLimit              = 20
	defaultHTTPTimeOut        = 10 * time.Second
	defaultSyncTimeout        = 1 * time.Hour
	defaultGoRequestRetry     = 3
	defaultGoRequestRetryTime = 1 * time.Second
)
