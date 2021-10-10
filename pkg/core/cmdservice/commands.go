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

type KillHardByPortCmd struct {
	ICmd
}

func (c *KillHardByPortCmd) GetBashScript() string {
	return `kill -9 $(lsof -t -i:${1})`
}

type KillByPortCmd struct {
	ICmd
}

func (c *KillByPortCmd) GetBashScript() string {
	return `kill $(lsof -t -i:${1})`
}
