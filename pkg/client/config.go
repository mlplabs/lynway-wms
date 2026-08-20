package client

type Config struct {
	BaseURL string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	Token   string `env:"TOKEN" envDefault:"12345"`
}
