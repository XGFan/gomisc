package mdns

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
)

func NsLookUp(name string) {
	ns := []string{
		"223.5.5.5",       //alidns
		"180.76.76.76",    //baidu dns
		"114.114.114.114", //114
		"8.8.4.4",         //google
		"1.0.0.1",         //cf
		"1.2.4.8",         //cnnic
		"119.29.29.29",    //dnspod
		"202.103.24.68",   //hb telecom
		"202.103.44.150",  //hb telecom
	}
	fmt.Println(ns)
	config := dns.ClientConfig{Port: "53", Timeout: 5}
	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), dns.TypeA)
	m.RecursionDesired = true

	strings := make([]string, 0)
	for _, n := range ns {
		r, _, err := c.Exchange(m, net.JoinHostPort(n, config.Port))
		if r == nil {
			log.Printf("*** error: %s\n", err.Error())
			continue
		}

		if r.Rcode != dns.RcodeSuccess {
			log.Printf("%v \n", r.Rcode)
			//log.Printf(" *** invalid answer name %s after MX query for %s\n", os.Args[1], os.Args[1])
			continue
		}
		// Stuff must be in the answer section
		for _, a := range r.Answer {
			if dns.Type(a.Header().Rrtype).String() == "A" {
				x := a.(*dns.A)
				//fmt.Printf("%v\n", a)
				//fmt.Println(reflect.TypeOf(a))
				//fmt.Println(x.A)
				//fmt.Printf("%s\n", x.A)
				//fmt.Printf("%s\t%s\n", n, x.A)
				flag := true
				for _, s := range strings {
					if s == x.A.String() {
						flag = false
						break
					}
				}
				if flag {
					strings = append(strings, x.A.String())
				}
				//fmt.Printf("%s\n", a.String())
			}
		}
	}
	for _, s := range strings {
		fmt.Println(s)
	}
}
