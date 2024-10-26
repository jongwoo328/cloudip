package ip

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

	binaryIP := ipToBinary(parsedIP, 128) // 128비트로 변환하되, CIDR 마스크 크기로 조절됨

	node := tree
	for i := 0; i < len(binaryIP); i++ {
		if node.IsLeaf {
			fmt.Println("Matched CIDR:", node.CIDR)
			return true
		}
		bit := binaryIP[i]
		node = node.Children[bit]
		if node == nil {
			return false // No matching CIDR
		}
	}

	fmt.Println("Matched CIDR:", node.CIDR)
	return node.IsLeaf
}

// Convert IP to binary string
func ipToBinary(ip net.IP, maskSize int) []byte {
	if ip.To4() != nil {
		ip = ip.To4() // 32-bit for IPv4
		maskSize = min(maskSize, 32)
	} else {
		ip = ip.To16() // 128-bit for IPv6
		maskSize = min(maskSize, 128)
	}

	binary := make([]byte, 0, maskSize)
	for _, octet := range ip {
		for i := 7; i >= 0 && len(binary) < maskSize; i-- {
			bit := (octet >> i) & 1
			binary = append(binary, '0'+byte(bit))
		}
	}
	return binary
}
