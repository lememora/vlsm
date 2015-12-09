/*
Automate the process of creating VLSM subnets
*/
package main

import (
  "fmt"
  "log"
  "math"
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
  hostBits := len(fmt.Sprintf("%b", numberOfHosts - 1)) // 0…numberOfHosts-1
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
  subnet.broadcast = CalcAddress(network.IP, subnet.poolSize + 1)
  subnet.poolRange[0] = CalcAddress(network.IP, 1)
  subnet.poolRange[1] = CalcAddress(network.IP, subnet.poolSize)

  return &subnet
}

func CalcBoundary(network *net.IPNet) net.IP {
  mask := network.Mask
  ones, bits := mask.Size()
  return CalcAddress(network.IP, uint32(math.Pow(2, float64(bits - ones))) - 1)
}

func NetworkHasAddress(network *net.IPNet, address net.IP) bool {
  boundary := CalcBoundary(network)
  boundaryInt := binary.BigEndian.Uint32(boundary.To4())
  addressInt := binary.BigEndian.Uint32(address.To4())
  return addressInt <= boundaryInt
}

func CalcVLSM(network *net.IPNet, subnetParams []SubnetParams) (subnets []Subnet, valid bool) {
  subnets = []Subnet{}
  valid = true
  nextNetwork := &net.IPNet{IP: network.IP, Mask: network.Mask}
  
  for i:= 0; i < len(subnetParams); i++ {
    params := subnetParams[i]
    numberOfHosts := (params.size + 2) // +(network+broadcast)
    // type_ := params.type_
    // 
    // // if(type_==61) {
    // //   
    // // }
    
    subnet := CalcSubnet(*nextNetwork, numberOfHosts)
    valid = NetworkHasAddress(network, subnet.broadcast)
    if valid {
      subnets = append(subnets, *subnet)
      /* next available network after subnetting */
      nextNetwork.IP = CalcAddress(subnet.broadcast, 1)
      nextNetwork.Mask = CalcMask(subnet.network.Mask, numberOfHosts)
    }    
  }
  
  return subnets, valid
}

func main() {
  /* Ask for parameters */
  
  // networkParams := NetworkParams{} // empty
  // networkParams := NetworkParams{"172.16.0.0/16", uint32(4)} // test
  networkParams := NetworkParams{"192.168.1.0/24", uint32(4)} // test

  network := AskForNetwork(networkParams)
  numberOfSubnets := int(AskForNumberOfSubnets(networkParams))

  // subnetParams := make([]SubnetParams, numberOfSubnets) // empty
  subnetParams := []SubnetParams{ // test
    // SubnetParams{113, 60},
    // SubnetParams{97,  60},
    // SubnetParams{61,  60},
    // SubnetParams{29,  60},
    SubnetParams{97, 60},
    SubnetParams{61, 60},
    SubnetParams{29, 60},
    SubnetParams{13, 60},
  }

  for i:= 0; i < numberOfSubnets; i++ {
    AskForSubnetParams(&subnetParams[i], i)
  }
  sort.Sort(SubnetParamsSort(subnetParams))

  /* Calculate Subnets */

  subnets, valid := CalcVLSM(network, subnetParams)
  
  if !valid {
    log.Fatal(fmt.Errorf("Network not big enough"))
  }

  /* Debug Subnets */

  fmt.Println("===== DEBUG: SUBNETS =====")
  for i:= 0; i < len(subnets); i++ {
    fmt.Printf("--- Subnet #%d ---\n", i)
    fmt.Printf("\tnetwork = %v\n", subnets[i].network)
    fmt.Printf("\tdotted mask = %s\n", subnets[i].dottedMask)
    fmt.Printf("\tbroadcast = %v\n", subnets[i].broadcast)
    fmt.Printf("\tpool size = %d\n", subnets[i].poolSize)
    fmt.Printf("\tpool range = FROM: %v  TO: %v\n", subnets[i].poolRange[0], subnets[i].poolRange[1])
  }
}