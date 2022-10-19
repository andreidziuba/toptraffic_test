package flags

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	. "github.com/andreidziuba/toptraffic_test/pkg/structures"
	flag "github.com/spf13/pflag"
)

func FlagsParse() (*int, *[]IPORT) {
	port_ref := flag.IntP("port", "p", -1, "port: 1-65535")
	advertisingPartnersSlice := flag.StringSliceP("dsp", "d", make([]string, 0), "ip:port,ip:port,... [1-10]")

	flag.Parse()
	if *port_ref < 1 || *port_ref > 65535 {
		fmt.Println("Port (flag -p) должен быть в пределах 1-65535", *port_ref)
		os.Exit(2)

	}
	advertisingPartners := parseAdvertisingPartners(advertisingPartnersSlice)

	return port_ref, advertisingPartners
}

// Парсим адреса рекламных партнёров
func parseAdvertisingPartners(advertisingPartnersSlice *[]string) *[]IPORT {
	advertisingPartners := make([]IPORT, 0, 10)
	for _, apString := range *advertisingPartnersSlice {
		apSplitIpPort := strings.Split(apString, ":")
		apPort, err := strconv.ParseUint(strings.Trim(apSplitIpPort[1], " "), 10, 64)
		if err != nil || apPort < 1 || apPort > 65535 {
			fmt.Println("port не явлется числом от 1 до 65535: ", apString)
			os.Exit(3)
		}
		apParsedIp := net.ParseIP(strings.Trim(apSplitIpPort[0], " "))
		if apParsedIp == nil {
			fmt.Println("Неверный ip", apSplitIpPort[0])
			os.Exit(3)
		}
		advertisingPartners = append(advertisingPartners, IPORT{IP: apParsedIp, Port: uint16(apPort)})
	}
	if len(advertisingPartners) > 10 {
		fmt.Println("Рекламных партнёров больше 10")
	}
	if len(advertisingPartners) < 1 {
		fmt.Println("Рекламных партнёров меньше 1")
		os.Exit(3)
	}
	return &advertisingPartners
}
