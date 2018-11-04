package pureglimpse

var WorkingDir = "./data/"

type Control struct {
	// Holds references to the other classes
	fetch   Fetcher
	reverse Reverser
}

func New() Control {
	c := Control{
		NewFetcher(),
		NewReverser(),
	}
	return c
}

func (c *Control) RunForever() {
	return
}
