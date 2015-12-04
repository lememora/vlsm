package main

import (
  "fmt"
  "log"
  "net"
  "strconv"
)

type Subnet struct {
  network net.IPNet
}

func main() {
  var arg, defaultArg string
  defaultArg = "192.168.0.0/16" /* default class C */
  fmt.Printf("Enter IPv4 network address in CIDR format (%s): ", defaultArg)
	n, err := fmt.Scanln(&arg)
  if n == 0 {
    arg = defaultArg
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }
  ip, net, err := net.ParseCIDR(arg)
  if err != nil {
    log.Fatal(fmt.Errorf("ParseCIDR(%q) = %v, %v", arg, ip, net))
  }

  defaultArg = "1"
  fmt.Printf("Enter the number of subnets (%s): ", defaultArg)
  n, err = fmt.Scanln(&arg)
  if n == 0 {
    arg = defaultArg
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }
  subnets, err := strconv.Atoi(arg)
  if err != nil {
    log.Fatal(fmt.Errorf("Atoi(%q) = %v", arg, subnets))
  }

  fmt.Println(ip)
  fmt.Println(net)
  fmt.Println(subnets)
}