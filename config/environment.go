package config

type Environment string

const (
	Development Environment = "DEV"
	Production  Environment = "PROD"
)

func (e Environment) String() string {
	return string(e)
}
