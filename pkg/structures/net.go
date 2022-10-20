package structures

import (
	"fmt"
	"net"
	"strconv"
)

type IPORT struct {
	IP   net.IP
	Port uint16
}

func (iport *IPORT) to_string() string {
	return fmt.Sprintf("%s:%s", iport.IP.String(), strconv.FormatUint(uint64(iport.Port), 10))
}

func (iport *IPORT) To_url(a string) string {
	if a != "" {
		return fmt.Sprintf("http://%s/%s", iport.to_string(), a)
	}
	return fmt.Sprintf("http://%s", iport.to_string())
}
