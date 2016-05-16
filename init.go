package multinet

func Init(ic, mc int) {
	if ic != 0 {
		initTCPCount = ic
	}
	if mc != 0 && mc > ic {
		maxTCPCount = ic
	}
}
