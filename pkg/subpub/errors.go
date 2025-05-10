package subpub

import "errors"

var (
	ErrorSubPubIsClosed  = errors.New("subpub is closed")
	ErrorChannelOverflow = errors.New("channel is full, cannot publish new msg")
)
