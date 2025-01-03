package common

const AppName = "cloudip"

type CloudIpFlag struct {
	Delimiter string
	Version   bool
	Format    string
	Header    bool
}

var Flags = CloudIpFlag{}
