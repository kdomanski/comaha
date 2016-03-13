package main

type userDB interface {
	AddPayload(id, sha1, sha256 string, size int64, version payloadVersion) error
	DeletePayload(id, channel string) error
	AttachPayloadToChannel(id, channel string) error
	GetNewerPayload(currentVersion payloadVersion, channel string) (*payload, error)
	PayloadExists(id string) bool

	ListImages(channel string) ([]payload, error)
	ListChannels() ([]string, error)
	GetChannelForceDowngrade(channel string) (bool, error)
	SetChannelForceDowngrade(channel string, value bool) error

	GetEvents() ([]Event, error)
	LogEvent(client string, evType, evResult int) error

	Close() error
}
