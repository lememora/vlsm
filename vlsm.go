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
	fmt.Scanln(&networkAddress)
  if len(networkAddress) == 0 {
    networkAddress = "192.168.0.0/16"
  }
  ip, net, err := net.ParseCIDR(networkAddress)
  if err != nil {
    log.Fatal(fmt.Errorf("ParseCIDR(%q) = %v, %v", networkAddress, ip, net))
  }
  fmt.Println(networkAddress)
}