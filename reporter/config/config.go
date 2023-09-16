package config

type Config struct {
	Input    string
	Output   string
	Root     string
	All      bool
	Cutlines *Cutlines
}
type Cutlines struct {
	Safe    float64
	Warning float64
}
