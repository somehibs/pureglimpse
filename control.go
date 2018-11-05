package pureglimpse

var WorkingDir = "./data/"

type Control struct {
	// Holds references to the other classes
	fetch   Fetcher
	reverse Reverser
	scan    Scanner
}

func New() Control {
	c := Control{
		NewFetcher(),
		NewReverser(),
		NewScanner(),
	}
	return c
}

func (c *Control) RunForever() {
	go c.fetch.StreamList(c.reverse.AppStream, 500)
	go c.reverse.StreamAppsForever(c.scan.ScanChan)
	c.scan.ScanAppsForever()
	return
}
