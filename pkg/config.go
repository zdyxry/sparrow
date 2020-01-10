package pkg

type Config struct {
	Raft    Raft    `toml:"raft"`
	Network Network `toml:"network"`
	Ntp     Ntp     `toml:"ntp"`
}
type Peers struct {
	ID   string `toml:"id"`
	Addr string `toml:"addr"`
}
type Raft struct {
	ID    string  `toml:"id"`
	Bind  string  `toml:"bind"`
	Peers []Peers `toml:"peers"`
}
type Network struct {
	Vip       string `toml:"vip"`
	Interface string `toml:"interface"`
}
type Ntp struct {
	Mode    string   `toml:"mode"`
	Servers []string `toml:"servers"`
}

