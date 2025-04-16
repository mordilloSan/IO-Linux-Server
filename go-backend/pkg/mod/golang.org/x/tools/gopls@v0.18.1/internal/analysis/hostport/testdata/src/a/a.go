package a

import (
	"fmt"
	"net"
)

func direct(host string, port int, portStr string) {
	// Dial, directly called with result of Sprintf.
	net.Dial("tcp", fmt.Sprintf("%s:%d", host, port)) // want `address format "%s:%d" does not work with IPv6`

	net.Dial("tcp", fmt.Sprintf("%s:%s", host, portStr)) // want `address format "%s:%s" does not work with IPv6`
}

// port is a constant:
var addr4 = fmt.Sprintf("%s:%d", "localhost", 123) // want `address format "%s:%d" does not work with IPv6 \(passed to net.Dial at L39\)`

func indirect(host string, port int) {
	// Dial, addr is immediately preceding.
	{
		addr1 := fmt.Sprintf("%s:%d", host, port) // want `address format "%s:%d" does not work with IPv6.*at L22`
		net.Dial("tcp", addr1)
	}

	// DialTimeout, addr is in ancestor block.
	addr2 := fmt.Sprintf("%s:%d", host, port) // want `address format "%s:%d" does not work with IPv6.*at L28`
	{
		net.DialTimeout("tcp", addr2, 0)
	}

	// Dialer.Dial, addr is declared with var.
	var dialer net.Dialer
	{
		var addr3 = fmt.Sprintf("%s:%d", host, port) // want `address format "%s:%d" does not work with IPv6.*at L35`
		dialer.Dial("tcp", addr3)
	}

	// Dialer.Dial again, addr is declared at package level.
	dialer.Dial("tcp", addr4)
}
