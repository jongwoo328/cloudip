package util

import (
	"fmt"
	"net"
)

// CIDRTree CIDR Tree structure
type CIDRTree struct {
	Children map[byte]*CIDRTree // Branch to each bit
	IsLeaf   bool               // Whether the node is end of the CIDR
	CIDR     string             // Save CIDR string if leaf node
}

// NewCIDRTree Create new CIDR tree
func NewCIDRTree() *CIDRTree {
	return &CIDRTree{
		Children: make(map[byte]*CIDRTree),
	}
}

// AddCIDR Add CIDR to tree
func (tree *CIDRTree) AddCIDR(cidr string) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Printf("Invalid CIDR: %s, error: %v\n", cidr, err)
		return // Invalid CIDR
	}

	maskSize, _ := ipNet.Mask.Size()
	binaryIP := ipToBinary(ipNet.IP, maskSize)

	node := tree
	for i := 0; i < maskSize; i++ {
		bit := binaryIP[i]
		if node.Children[bit] == nil {
			node.Children[bit] = NewCIDRTree()
		}
		node = node.Children[bit]
	}
	node.IsLeaf = true
	node.CIDR = cidr
}

// Match Verify that the IP belongs to one of the CIDRs in the CIDR tree
func (tree *CIDRTree) Match(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return tree.MatchParsedIP(parsedIP)
}

// MatchParsedIP verifies that the parsed IP belongs to one of the CIDRs in the tree.
func (tree *CIDRTree) MatchParsedIP(parsedIP net.IP) bool {
	if parsedIP == nil {
		return false
	}

	ipBytes := parsedIP.To4()
	if ipBytes == nil {
		ipBytes = parsedIP.To16()
		if ipBytes == nil {
			return false
		}
	}

	node := tree
	for _, octet := range ipBytes {
		for bitIndex := 7; bitIndex >= 0; bitIndex-- {
			if node.IsLeaf {
				return true
			}

			bit := (octet >> bitIndex) & 1
			node = node.Children[bit]
			if node == nil {
				return false // No matching CIDR
			}
		}
	}

	return node.IsLeaf
}

// Convert IP to binary string
func ipToBinary(ip net.IP, maskSize int) []byte {
	if ip.To4() != nil {
		ip = ip.To4() // 32-bit for IPv4
		maskSize = min(maskSize, net.IPv4len*8)
	} else {
		ip = ip.To16() // 128-bit for IPv6
		maskSize = min(maskSize, net.IPv6len*8)
	}

	binary := make([]byte, 0, maskSize)
	for _, octet := range ip {
		for i := 7; i >= 0 && len(binary) < maskSize; i-- {
			bit := (octet >> i) & 1
			binary = append(binary, bit)
		}
	}
	return binary
}
