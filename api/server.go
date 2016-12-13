package api

type Server interface {
	Version() string
	BuildTime() string
}
