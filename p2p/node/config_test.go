package node

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToMultiaddrs_ValidAddresses(t *testing.T) {
	/// Given
	addresses := []string{
		"/ip6/::/tcp/10015",
		"/ip4/0.0.0.0/tcp/10015",
		"/ip4/192.168.0.13/tcp/80/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	}

	/// When
	multiaddrs, err := ToMultiaddrs(addresses)

	/// Then
	assert.NoErrorf(t, err, "Error converting addresses to Multiaddrs: %v", err)

	addrIp6Protocols := multiaddrs[0].Protocols()
	assert.Lenf(t, addrIp6Protocols, 2, "Expected 2 protocols, got %d", len(addrIp6Protocols))
	assert.Equal(t, "ip6", addrIp6Protocols[0].Name)
	assert.Equal(t, "tcp", addrIp6Protocols[1].Name)

	addrIp4Protocols := multiaddrs[1].Protocols()
	assert.Lenf(t, addrIp4Protocols, 2, "Expected 2 protocols, got %d", len(addrIp4Protocols))
	assert.Equal(t, "ip4", addrIp4Protocols[0].Name)
	assert.Equal(t, "tcp", addrIp4Protocols[1].Name)

	addrIp4P2pProtocols := multiaddrs[2].Protocols()
	assert.Lenf(t, addrIp4P2pProtocols, 3, "Expected 3 protocols, got %d", len(addrIp4P2pProtocols))
	assert.Equal(t, "ip4", addrIp4P2pProtocols[0].Name)
	assert.Equal(t, "tcp", addrIp4P2pProtocols[1].Name)
	assert.Equal(t, "p2p", addrIp4P2pProtocols[2].Name)
}

func TestToMultiaddrs_InvalidAddresses(t *testing.T) {
	/// Given
	addresses := []string{"gibberish"}

	/// When
	_, err := ToMultiaddrs(addresses)

	/// Then
	assert.Error(t, err, "Expected error converting invalid addresses to Multiaddrs")
}

func TestToPeerAddrInfo_ValidAddresses(t *testing.T) {
	/// Given
	addresses := []string{
		"/ip4/10.20.30.40/tcp/443/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
		"/ip4/192.168.0.13/tcp/80/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	}

	/// When
	addrInfo, err := ToAddrInfo(addresses)

	/// Then
	assert.NoErrorf(t, err, "Error converting addresses to AddrInfo: %v", err)

	assert.Equal(t, addrInfo[0].ID.String(), "QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	assert.Equal(t, addrInfo[0].Addrs[0].String(), "/ip4/10.20.30.40/tcp/443")

	assert.Equal(t, addrInfo[1].ID.String(), "QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb")
	assert.Equal(t, addrInfo[1].Addrs[0].String(), "/ip4/192.168.0.13/tcp/80")
}

func TestToPeerAddrInfo_InvalidAddress(t *testing.T) {
	/// Given
	addresses := []string{
		"/ip4/10.20.30.40/tcp/443/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
		"avocado",
	}

	/// When
	_, err := ToAddrInfo(addresses)

	/// Then
	assert.Errorf(t, err, "Expected error converting addresses to AddrInfo: %v", err)
}

func TestToPeerAddrInfo_NoPeerIdAddress(t *testing.T) {
	/// Given
	addresses := []string{
		"/ip4/10.20.30.40/tcp/443/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N",
		"/ip4/192.168.0.13/tcp/80",
	}

	/// When
	_, err := ToAddrInfo(addresses)

	/// Then
	assert.Errorf(t, err, "Expected error converting addresses to AddrInfo: %v", err)
}
func TestToDiscoveryMethod_EmptyMethodsSlice(t *testing.T) {
	/// Given
	methods := []string{}

	/// When
	discoveryMethods, err := ToDiscoveryMethod(methods)

	/// Then
	assert.NoErrorf(t, err, "Error converting methods to DiscoveryMethod: %v", err)
	assert.Len(t, discoveryMethods, 1)
	assert.Containsf(t, discoveryMethods, DiscoveryMethodNone, "Expected discovery method to contain %v", DiscoveryMethodNone)
}

func TestToDiscoveryMethod_InvalidMethod(t *testing.T) {
	/// Given
	methods := []string{"invalid"}

	/// When
	discoveryMethods, err := ToDiscoveryMethod(methods)

	/// Then
	assert.Error(t, err)
	assert.Nil(t, discoveryMethods)
}

func TestToDiscoveryMethod_VariousMethodsAndNone(t *testing.T) {
	/// Given
	methods := []string{"mdns", "none"}

	/// When
	discoveryMethods, err := ToDiscoveryMethod(methods)

	/// Then
	assert.NoErrorf(t, err, "Error converting methods to DiscoveryMethod: %v", err)
	assert.Len(t, discoveryMethods, 1)
	assert.Containsf(t, discoveryMethods, DiscoveryMethodNone, "Expected discovery method to contain %v", DiscoveryMethodNone)
}

func TestToDiscoveryMethod_VariousRepeatedMethods(t *testing.T) {
	/// Given
	methods := []string{"mdns", "mdns", "mdns", "dht", "mdns", "mdns", "dht"}

	/// When
	discoveryMethods, err := ToDiscoveryMethod(methods)

	/// Then

	assert.NoErrorf(t, err, "Error converting methods to DiscoveryMethod: %v", err)
	assert.Len(t, discoveryMethods, 2)
	assert.Containsf(t, discoveryMethods, DiscoveryMethodDHT, "Expected discovery method to contain %v", DiscoveryMethodDHT)
	assert.Containsf(t, discoveryMethods, DiscoveryMethodMDNS, "Expected discovery method to contain %v", DiscoveryMethodMDNS)
}
