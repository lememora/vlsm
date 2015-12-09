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
)

type NetworkParams struct {
  networkAddress string
  numberOfSubnets uint32
}

type Subnet struct {
  network net.IPNet
  mask net.IP
  broadcast net.IP
  poolSize uint32
  poolRange [2]net.IP
}

type SubnetParams struct {
  size uint32
  distribution byte
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

func enterNetwork(p NetworkParams) *net.IPNet {
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

func enterNumberOfSubnets(p NetworkParams) uint32 {
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

func enterSubnetSize(p *SubnetParams) {
  var arg string
  argDefault := "2"
  fmt.Printf("Enter subnet size (%s): ", argDefault)
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

func enterSubnetDistribution(p *SubnetParams) {
  // options := map[byte]string {60: "minimum", 61: "balanced", 62: "maximum"}
  // optionKeys := make([]int, 0, len(options))
  // // contains(options, arg)
  var arg string
  argDefault := "<"
  fmt.Printf("Enter subnet distribution [<|=|>] (%s): ", argDefault)
  n, err := fmt.Scanln(&arg)

  if n == 0 {
    arg = argDefault
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }

  p.distribution = byte(arg[0])
}

func enterSubnetParams(p *SubnetParams, counter int) {
  if p.size == 0 {
    fmt.Printf("=== Subnet #%d ===\n", counter + 1)
    enterSubnetSize(p)    
  }
  if !(p.size >= 1 && p.size <= 2147483646) {
    log.Fatal(fmt.Errorf("Invalid subnet size = %d", p.size))
  }
  if p.distribution == 0 {
    enterSubnetDistribution(p)    
  }
  if !(p.distribution >= 60 && p.distribution <= 62) {
    log.Fatal(fmt.Errorf("Invalid subnet distribution = %d", p.distribution))
  }
}

func main() {
  networkParams := NetworkParams{"192.168.1.0/24", uint32(5)} // test
  // networkParams := NetworkParams{} // empty

  network := enterNetwork(networkParams)
  
  /* Calculate number of host bits available for the given network */
  ones, bits := network.Mask.Size()
  availableHostBits := bits - ones

  numberOfSubnets := int(enterNumberOfSubnets(networkParams))

  // subnetParams := make([]SubnetParams, numberOfSubnets) // empty
  subnetParams := []SubnetParams{ // test
    SubnetParams{50,  60},
    SubnetParams{150, 60},
    SubnetParams{10,  60},
    SubnetParams{5,   60},
    SubnetParams{30,  60},
  }

  for i:= 0; i < numberOfSubnets; i++ {
    enterSubnetParams(&subnetParams[i], i)
  }
  sort.Sort(SubnetParamsSort(subnetParams))

  /* Calculate number of host bits necessary based on the biggest subnet */
  requiredHostBits := int(math.Ceil(math.Log2(float64(subnetParams[0].size))))
  availableSubnetBits := availableHostBits - requiredHostBits
  if availableSubnetBits < 0 {
    log.Fatal(fmt.Errorf("Network not big enough"))
  }
  
  // subnets := make([]Subnet, 0)
  
  
  fmt.Println("=== DEBUG >>>")
  fmt.Printf("network = %v\n", network)
  fmt.Printf("network.Mask = %v\n", network.Mask)
  fmt.Printf("availableHostBits = %v\n", availableHostBits)
  fmt.Printf("numberOfSubnets = %v\n", numberOfSubnets)
  fmt.Printf("subnetParams = %v\n", subnetParams)
  fmt.Printf("requiredHostBits = %v\n", requiredHostBits)
  fmt.Println("<<< DEBUG ===")
  // fmt.Println(subnets)
}