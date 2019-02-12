package main

type rule struct {
	LocalSock  string `toml:"local"`
	RemoteSock string `toml:"remote"`
}
