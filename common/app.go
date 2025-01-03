package common

const AppName = "cloudip"

type CloudIpFlag struct {
	Delimiter string
	Format    string
	Header    bool
}

var Flags = CloudIpFlag{}
