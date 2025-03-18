package client

var (
	Host    string = "http://localhost:5555"
	Version string = "dev"
)

type Options struct {
	HostURL string
	ScreenW int
	ScreenH int
	Version string
}
