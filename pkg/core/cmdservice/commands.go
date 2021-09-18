package cmdservice

type NetStatCmd struct {
	ICmd
}

func (c *NetStatCmd) GetBashScript() string {
	return `netstat -aon | grep "${1}"`
}

type NslookupCmd struct {
	ICmd
}

func (c *NslookupCmd) GetBashScript() string {
	return `nslookup "${1}"`
}
