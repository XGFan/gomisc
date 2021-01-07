package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"github.com/cloudflare/cloudflare-go"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var debug = false

type Ddns6Config struct {
	Provider string            `json:"provider"`
	Config   map[string]string `json:"config"`
	Ddns     []DdnsItem        `json:"ddns"`
}

type DdnsItem struct {
	Mac    string `json:"mac"`
	Domain string `json:"domain"`
}

func AddressIsType(flags net.Flags, t net.Flags) bool {
	return flags&t != 0
}

func AddressIsNotType(flags net.Flags, t net.Flags) bool {
	return !(flags&t != 0)
}

func IsULA(ip net.IP) bool {
	_, block, _ := net.ParseCIDR("fc00::/7")
	return block.Contains(ip)
}

func findIPv6() ([]net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	ip := make([]net.IP, 0)
	for _, i := range interfaces {
		byNameInterface, err := net.InterfaceByName(i.Name)
		if err != nil {
			return nil, err
		}
		addresses, err := byNameInterface.Addrs()
		if err != nil {
			return nil, err
		}
		//过滤掉未连接/回环/无地址的网卡
		if len(addresses) == 0 ||
			AddressIsType(i.Flags, net.FlagLoopback) ||
			AddressIsType(i.Flags, net.FlagPointToPoint) ||
			AddressIsNotType(i.Flags, net.FlagUp) {
			continue
		}
		ips := make([]*net.IPNet, 0)
		for _, v := range addresses {
			ipNet, ok := v.(*net.IPNet)
			//过滤掉本地地址,ipv4地址,ULA地址
			if !ok ||
				ipNet.IP.To4() != nil ||
				ipNet.IP.IsLinkLocalUnicast() ||
				ipNet.IP.IsLinkLocalMulticast() ||
				ipNet.IP.IsInterfaceLocalMulticast() ||
				IsULA(ipNet.IP) {
				continue
			}
			ips = append(ips, ipNet)
		}
		//如果没有任何满足要求的地址
		if len(ips) == 0 {
			continue
		}
		//log.Printf("Name : %v \n", i.Name)
		//log.Printf("Flag : %v \n", i.Flags)
		log.Println("Interface HardwareAddr: ", byNameInterface.HardwareAddr)
		for _, v := range ips {
			log.Printf("Interface Address: %v\n", v.IP)
			//	log.Printf("Interface Address IsGlobalUnicast: %v\n", v.IP.IsGlobalUnicast())
			//	log.Printf("Interface Address IsMulticast: %v\n", v.IP.IsMulticast())
			//	log.Printf("Interface Address IsUnspecified: %v\n", v.IP.IsUnspecified())
			//	log.Printf("Interface Address IsLoopback: %v\n", v.IP.IsLoopback())
			ip = append(ip, v.IP)
		}
	}
	return ip, nil
}

func findPrefix(ips []net.IP) []byte {
	if len(ips) == 0 {
		return nil
	}
	if len(ips) == 1 {
		return ips[0][0:8]
	}
	counter := make(map[string]int, 0)
	for _, ip := range ips {
		ipStr := hex.EncodeToString(ip[0:8])
		counter[ipStr] = counter[ipStr] + 1
	}
	max := 0
	ip := ""
	for k, v := range counter {
		if v > max {
			ip = k
		}
	}
	bytes, _ := hex.DecodeString(ip)
	return bytes
}

func EUI64(mac string) string {
	bytes := make([]byte, 0)
	strArray := make([]string, 8)
	split := strings.Split(mac, ":")
	copy(strArray, split[0:3])
	copy(strArray[3:5], []string{"FF", "FE"})
	copy(strArray[5:], split[3:6])
	for _, s := range strArray {
		bs, error := hex.DecodeString(s)
		if error != nil {
			log.Fatal(error)
		}
		bytes = append(bytes, bs...)
	}
	t := bytes[0] & (1 << 1)
	if t == 0 {
		bytes[0] = bytes[0] | (1 << 1)
	} else {
		bytes[0] = bytes[0] ^ (1 << 1)
	}
	return BytesToString(bytes)
}

func BytesToString(bytes []byte) string {
	strSlice := make([]string, 0)
	for i := 0; i < len(bytes); i = i + 2 {
		strSlice = append(strSlice, hex.EncodeToString(bytes[i:i+2]))
	}
	return strings.Join(strSlice, ":")
}

func main() {
	flag.Parse()
	file, err := getFile(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	config := getConfig(file)
	ipv6, e := findIPv6()
	if e != nil {
		log.Fatal(e)
	}
	prefix := findPrefix(ipv6)
	if config.Provider == "cf" {
		CF{}.updateDNS(config.Config, config.Ddns, prefix)
	}
}

type CF struct {
}

func assembleIPv6String(prefix []byte, mac string) string {
	prefixStr := BytesToString(prefix)
	eui64Str := EUI64(mac)
	log.Printf("From: [%s]:[%s]", prefixStr, mac)
	log.Printf("To: [%s]:[%s]", prefixStr, eui64Str)
	return prefixStr + ":" + eui64Str
}

func (c CF) updateDNS(config map[string]string, items []DdnsItem, prefix []byte) {
	api, err := cloudflare.New(config["key"], config["email"])
	if err != nil {
		log.Fatal(err)
	}
	zoneId, err := api.ZoneIDByName(config["zone"])
	if err != nil {
		log.Fatal(err)
	}
	if len(items) == 0 {
		log.Printf("%s item is empty", reflect.TypeOf(c))
		return
	}
	records, _ := api.DNSRecords(zoneId, cloudflare.DNSRecord{Type: "AAAA"})
	recordsMap := make(map[string]cloudflare.DNSRecord, 0)
	for _, record := range records {
		recordsMap[record.Name] = record
	}
	var wg sync.WaitGroup
	wg.Add(len(items))
	for _, ddns := range items {
		fullIp := assembleIPv6String(prefix, ddns.Mac)
		dnsRecord := cloudflare.DNSRecord{Type: "AAAA", Name: ddns.Domain, Proxied: false, Content: fullIp, TTL: 1}
		if record, ok := recordsMap[ddns.Domain]; ok {
			content := record.Content
			if content == fullIp {
				log.Printf("%s is Update [%s]", ddns.Domain, fullIp)
				wg.Done()
			} else {
				go func() {
					defer wg.Done()
					e := api.UpdateDNSRecord(zoneId,
						record.ID,
						dnsRecord)
					log.Printf("%s has Updated [%s]", ddns.Domain, fullIp)
					if e != nil {
						log.Println(e)
					}
				}()
			}
		} else {
			go func() {
				defer wg.Done()
				_, e := api.CreateDNSRecord(zoneId, dnsRecord)
				if e != nil {
					log.Println(e)
				}
				log.Printf("%s has Created [%s]", ddns.Domain, fullIp)
			}()
		}
	}
	wg.Wait()
}

func getConfig(file string) *Ddns6Config {
	bytes, e := ioutil.ReadFile(file)
	if e != nil {
		log.Fatal(e)
	}
	configs := &Ddns6Config{}
	e = json.Unmarshal(bytes, configs)
	if e != nil {
		log.Fatal(e)
	}
	return configs
}

func getFile(fileName string) (file string, e error) {
	if debug {
		log.Printf("get file %s", fileName)
	}
	dir, _ := os.Getwd()                 //get current dir
	join := filepath.Join(dir, fileName) //get relative path
	if fileExists(join) {
		file = join
	} else {
		if fileExists(fileName) {
			file = fileName
			if debug {
				log.Println("absolute yes")
			}
		} else {
			e = errors.New("file not found")
		}
	}
	return
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
