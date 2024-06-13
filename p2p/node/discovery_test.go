package node

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseDiscoveryMethod_ValidStrings(t *testing.T) {
	m, err := ParseDiscoveryMethod("mdns")
	assert.Nil(t, err)
	assert.Equal(t, DiscoveryMethodMDNS, m)

	m, err = ParseDiscoveryMethod("MDNS")
	assert.Nil(t, err)
	assert.Equal(t, DiscoveryMethodMDNS, m)

	m, err = ParseDiscoveryMethod("dht")
	assert.Nil(t, err)
	assert.Equal(t, DiscoveryMethodDHT, m)

	m, err = ParseDiscoveryMethod("DHT")
	assert.Nil(t, err)
	assert.Equal(t, DiscoveryMethodDHT, m)

	m, err = ParseDiscoveryMethod("none")
	assert.Nil(t, err)
	assert.Equal(t, DiscoveryMethodNone, m)

	m, err = ParseDiscoveryMethod("NONE")
	assert.Nil(t, err)
	assert.Equal(t, DiscoveryMethodNone, m)
}

func TestDiscoveryMethod_InvalidString(t *testing.T) {
	_, err := ParseDiscoveryMethod("invalid")
	assert.Equal(t, "not a valid discovery method: \"invalid\"", err.Error())
}

func TestConvertDiscoveryMethodToString(t *testing.T) {
	assert.Equal(t, "none", DiscoveryMethodNone.String())
	assert.Equal(t, "mdns", DiscoveryMethodMDNS.String())
	assert.Equal(t, "dht", DiscoveryMethodDHT.String())
}
