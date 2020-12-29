package main

import (
	"flag"
	"net"

	"fmt"

	geoip2 "github.com/oschwald/geoip2-golang"
)

var (
	argsInputIP = flag.String("ip", "", "client ip address")
)

func main() {
	flag.Parse()

	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ip := net.ParseIP(*argsInputIP)
	record, err := db.City(ip)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("ISO Country code: %v\n", record.Country.IsoCode)
	fmt.Printf("Country: %s\n", record.Country.Names["zh-CN"])
	fmt.Printf("%+v\n", record)
	//fmt.Printf("Subdivision: %v\n", record.Subdivisions)

}
