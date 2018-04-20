package bot

import (
	"strconv"
)

//server_port, local_port, remote_port
var rpc = `
[common]
server_addr = localhost
server_port = %v
http_proxy =

[rpc%v]
type = tcp
local_ip = localhost
local_port = %v
remote_port = %v
`

func parseInt(s string, v int) int {
	if s == "" {
		return v
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		i = v
	}
	return i
}
