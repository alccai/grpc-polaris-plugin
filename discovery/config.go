package discovery

type Config struct {
	Name string
}

type FactoryConfig struct {
	AddressList string   `yaml:"address_list"`
	Clients     []Client `yaml:"clients"`
}

type Client struct {
	Name string `yaml:"name"`
}
