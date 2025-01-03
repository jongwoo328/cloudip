package common

const AppName = "cloudip"

type CloudIpFlag struct {
	Delimiter string
	Format    string
	Header    bool
	Verbose   bool
}

var Flags = &CloudIpFlag{}
