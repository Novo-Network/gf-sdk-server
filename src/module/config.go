package module

type Config struct {
	Server ServerCfg `json:"server" toml:"server"`
}

type ServerCfg struct {
	PrivateKeyPath string `json:"private_key_path" toml:"private_key_path"`
	Host           string `json:"host" toml:"host"`
	ChainRpc       string `json:"chain_rpc" toml:"chain_rpc"`
	ChainId        string `json:"chain_id" toml:"chain_id"`
}
