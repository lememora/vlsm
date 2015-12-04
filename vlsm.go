/*
Automate the process of creating VLSM subnets
*/
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

type SubnetParams struct {
  size, distribution int
}

func enterNetwork() *net.IPNet {
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

  return net
}

func enterNumOfSubnets() int {
  var arg, defaultArg string
  defaultArg = "1"
  fmt.Printf("Enter the number of subnets (%s): ", defaultArg)
  n, err := fmt.Scanln(&arg)
  if n == 0 {
    arg = defaultArg
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }

  var numOfSubnets int
  numOfSubnets, err = strconv.Atoi(arg)
  if err != nil {
    log.Fatal(fmt.Errorf("Atoi(%q) = %v", arg, numOfSubnets))
  }
  if !(numOfSubnets >= 1 && numOfSubnets <= 4194304) {
    log.Fatal(fmt.Errorf("Invalid number of subnets = %d", numOfSubnets))
  }

  return numOfSubnets
}

func enterSubnetSize(counter int) int {
  var arg, defaultArg string
  defaultArg = "2"
  fmt.Printf("Enter subnet %d size (%s): ", counter, defaultArg)
  n, err := fmt.Scanln(&arg)
  if n == 0 {
    arg = defaultArg
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }

  var subnetSize int
  subnetSize, err = strconv.Atoi(arg)
  if err != nil {
    log.Fatal(fmt.Errorf("Atoi(%q) = %v", arg, subnetSize))
  }
  if !(subnetSize >= 2 && subnetSize <= 16777214) {
    log.Fatal(fmt.Errorf("Invalid subnet %d size = %d", counter, subnetSize))
  }

  return subnetSize
}

func enterSubnetDistribution(counter int) int {
  // distributionChoices := [3]string {"minimum","maximum","balanced"}
  // contains(distributionChoices, arg)
  var arg, defaultArg string
  defaultArg = "minimum"
  fmt.Printf("Enter subnet %d distribution [minimum|maximum|balanced] (%s): ", counter, defaultArg)
  n, err := fmt.Scanln(&arg)
  if n == 0 {
    // arg = defaultArg
    arg = "1" // TODO
  } else if err != nil {
    log.Fatal(fmt.Errorf("%s\n", err))
  }

  var distribution int
  distribution, err = strconv.Atoi(arg)
  if err != nil {
    log.Fatal(fmt.Errorf("Atoi(%q) = %v", arg, distribution))
  }
  if !(distribution >= 1 && distribution <= 3) {
    log.Fatal(fmt.Errorf("Invalid subnet %d distribution = %d", counter, distribution))
  }

  return distribution
}

func enterSubnetParams(counter int) *SubnetParams {
  return &SubnetParams{
    enterSubnetSize(counter),
    enterSubnetDistribution(counter),
  }
}

func main() {
  network := enterNetwork()
  numOfSubnets := enterNumOfSubnets()

  subnetParams := make([]SubnetParams, 0)
  for i:= 1; i <= numOfSubnets; i++ {
    subnetParams = append(subnetParams, *enterSubnetParams(i))
  }
  
  fmt.Println(network)
  fmt.Println(numOfSubnets)
  fmt.Println(subnetParams)
}