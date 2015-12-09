/*
Automate the process of creating VLSM subnets
*/
package main

import (
  "fmt"
  "log"
  // "math"
  "net"
  "sort"
  "strconv"
  "strings"
	"encoding/binary"
)

type NetworkParams struct {
  networkAddress string
  numberOfSubnets uint32
}

type Subnet struct {
  network net.IPNet
  dottedMask string
  broadcast net.IP
  poolSize uint32
  poolRange [2]net.IP
}

type SubnetParams struct {
  size uint32
  type_ byte
}

type SubnetParamsSort []SubnetParams

func (s SubnetParamsSort) Len() int {
  return len(s)
}

func (s SubnetParamsSort) Swap(i, j int) {
  s[i], s[j] = s[j], s[i]
}

func (s SubnetParamsSort) Less(i, j int) bool {
  return s[i].size > s[j].size
}

func AskForNetwork(p NetworkParams) *net.IPNet {
  if len(p.networkAddress) == 0 {
    argDefault := "10.0.0.0/8"
    fmt.Printf("Enter IPv4 network address in CIDR format (%s): ", argDefault)
    n, err := fmt.Scanln(&p.networkAddress)
    if n == 0 {
      p.networkAddress = argDefault
    } else if err != nil {
      log.Fatal(fmt.Errorf("%s\n", err))
    }
  }

  ip, net, err := net.ParseCIDR(p.networkAddress)
  if err != nil {
    log.Fatal(fmt.Errorf("ParseCIDR(%q) = %v, %v", p.networkAddress, ip, net))
  }

  return net
}

func AskForNumberOfSubnets(p NetworkParams) uint32 {
  if p.numberOfSubnets == 0 {
    var arg string
    argDefault := "1"
    fmt.Printf("Enter the number of subnets (%s): ", argDefault)
    n, err := fmt.Scanln(&arg)
    if n == 0 {
      arg = argDefault
    } else if err != nil {
      log.Fatal(fmt.Errorf("%s\n", err))
    }

    n, err  = strconv.Atoi(arg)
    if err != nil {
      log.Fatal(fmt.Errorf("Atoi(%q) = %v", arg, n))
    }

    p.numberOfSubnets = uint32(n)    
  }

  if !(p.numberOfSubnets >= 1 && p.numberOfSubnets <= 2147483648) {
    log.Fatal(fmt.Errorf("Invalid number of subnets = %d", p.numberOfSubnets))
  }

  return p.numberOfSubnets
}

func AskForSubnetSize(p *SubnetParams) {
  var arg string
  argDefault := "1"
  fmt.Printf("Enter subnet size or “number of assignable addresses” (%s): ", argDefault)
  n, err := fmt.Scanln(&arg)
  if n == 0 {
    arg = argDefault
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }

  n, err  = strconv.Atoi(arg)
  if err != nil {
    log.Fatal(fmt.Errorf("Atoi(%q) = %v", arg, n))
  }

  p.size = uint32(n)
}

func AskForSubnetType(p *SubnetParams) {
  var arg string
  argDefault := "<"
  fmt.Printf("Enter subnet type minimum|balanced|maximum [<|=|>] (%s): ", argDefault)
  n, err := fmt.Scanln(&arg)

  if n == 0 {
    arg = argDefault
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }

  p.type_ = byte(arg[0])
}

func AskForSubnetParams(p *SubnetParams, counter int) {
  if p.size == 0 {
    fmt.Printf("=== Subnet #%d ===\n", counter + 1)
    AskForSubnetSize(p)    
  }
  if !(p.size >= 1 && p.size <= 2147483646) {
    log.Fatal(fmt.Errorf("Invalid subnet size = %d", p.size))
  }
  if p.type_ == 0 {
    AskForSubnetType(p)    
  }
  if !(p.type_ >= 60 && p.type_ <= 62) {
    log.Fatal(fmt.Errorf("Invalid subnet type_ = %d", p.type_))
  }
}

func CalcPoolSize(numberOfHosts uint32) uint32 {
  hostBits := len(fmt.Sprintf("%b", numberOfHosts - 1))
  i, err := strconv.ParseInt(strings.Repeat("1", hostBits), 2, 32)
  if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }
  return uint32(i) - 1
}

func CalcAddress(ip net.IP, numberOfHosts uint32) net.IP {
  n := binary.BigEndian.Uint32(ip.To4()) + numberOfHosts
  return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

func CalcMask(mask net.IPMask, numberOfHosts uint32) net.IPMask {
  ones, bits := mask.Size()
  availableHostBits := bits - ones
  hostBitsNeeded := len(fmt.Sprintf("%b", numberOfHosts))
  networkBits := availableHostBits - hostBitsNeeded

  if networkBits < 0 {
    log.Fatal(fmt.Errorf("Network not big enough"))
  }
  
  return net.CIDRMask(ones + networkBits, 32)
}

func CalcSubnet(network net.IPNet, numberOfHosts uint32) *Subnet {
  subnet := Subnet{}

  subnet.network = network
  subnet.network.Mask = CalcMask(network.Mask, numberOfHosts)
  m := subnet.network.Mask
  subnet.dottedMask = fmt.Sprintf("%d.%d.%d.%d", m[0], m[1], m[2], m[3])
  subnet.poolSize = CalcPoolSize(numberOfHosts)
  subnet.broadcast = CalcAddress(network.IP, subnet.poolSize + 1) // wrong
  subnet.poolRange[0] = CalcAddress(network.IP, 1)
  subnet.poolRange[1] = CalcAddress(network.IP, subnet.poolSize)

  return &subnet
}

func main() {
  /* Ask for parameters */
  
  networkParams := NetworkParams{"172.16.0.0/16", uint32(4)} // test
  // networkParams := NetworkParams{} // empty

  network := AskForNetwork(networkParams)  
  numberOfSubnets := int(AskForNumberOfSubnets(networkParams))

  // subnetParams := make([]SubnetParams, numberOfSubnets) // empty
  subnetParams := []SubnetParams{ // test
    SubnetParams{50, 60},
    SubnetParams{25, 60},
    SubnetParams{10, 60},
    SubnetParams{2,  60},
  }

  for i:= 0; i < numberOfSubnets; i++ {
    AskForSubnetParams(&subnetParams[i], i)
  }
  sort.Sort(SubnetParamsSort(subnetParams))

  /* Calculate Subnets */

  subnets := []Subnet{}
  nextNetwork := network
  
  for i:= 0; i < numberOfSubnets; i++ {
    params := subnetParams[i]
    numberOfHosts := (params.size + 2) // +(network+broadcast)
    subnet := CalcSubnet(*nextNetwork, numberOfHosts)
    subnets = append(subnets, *subnet)
    
    /* next available network after subnetting */
    nextNetwork.IP = CalcAddress(subnet.broadcast, 1)
    nextNetwork.Mask = CalcMask(subnet.network.Mask, numberOfHosts)
  }

  /* Debug Subnets */

  fmt.Println("===== DEBUG: SUBNETS =====")
  for i:= 0; i < numberOfSubnets; i++ {
    fmt.Printf("--- Subnet #%d ---\n", i)
    fmt.Printf("\tnetwork = %v\n", subnets[i].network)
    fmt.Printf("\tdotted mask = %s\n", subnets[i].dottedMask)
    fmt.Printf("\tbroadcast = %v\n", subnets[i].broadcast)
    fmt.Printf("\tpool size = %d\n", subnets[i].poolSize)
    fmt.Printf("\tpool range = FROM: %v  TO: %v\n", subnets[i].poolRange[0], subnets[i].poolRange[1])
  }
}