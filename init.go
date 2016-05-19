package multinet

// Init reset the first tcp count and max tcp count for one group
func Init(ic, mc int) {
	if ic != 0 {
		initTCPCount = ic
	}
	if mc != 0 && mc > ic {
		maxTCPCount = ic
	}
}
