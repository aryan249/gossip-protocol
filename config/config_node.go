package config

type P2PConfig struct {
	PrivateKey             string
	ListenAddresses        []string
	BootstrapNodeAddresses []string
	DiscoveryMethods       []string
	EnableSync             bool
	SendToPerpApi          bool
}

func (v *viperConfig) ReadP2PConfig() P2PConfig {
	return P2PConfig{
		PrivateKey:             v.GetString("node_details.private_key"),
		ListenAddresses:        v.GetStringSlice("p2p_config.listen_addresses"),
		BootstrapNodeAddresses: v.GetStringSlice("p2p_config.bootstrap_nodes"),
		DiscoveryMethods:       v.GetStringSlice("p2p_config.discovery_methods"),
		EnableSync:             v.GetBool("p2p_config.enable_sync"),
		SendToPerpApi:          v.GetBool("p2p_config.send_to_perp_api"),
	}
}