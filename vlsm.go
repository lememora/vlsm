package main

import (
  "fmt"
  "log"
  "net"
)

type Subnet struct {
  network net.IPNet
}

func main() {  
  fmt.Printf("Enter IPv4 network address in CIDR format (192.168.0.0/16): ")
  var networkAddress string
	n, err := fmt.Scanln(&networkAddress)
  if n == 0 {
    networkAddress = "192.168.0.0/16" /* default class C */
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }
  ip, net, err := net.ParseCIDR(networkAddress)
  if err != nil {
    log.Fatal(fmt.Errorf("ParseCIDR(%q) = %v, %v", networkAddress, ip, net))
  }
  fmt.Println(networkAddress)
}