package def

type Person struct {
	Name        string      `fuzz:"length(8,32)"`
	Age         int         `fuzz:"range(1,100)"`
	Address     string      `fuzz:"length(20,60)"`
	Education   []Education `fuzz:"length(4)"`
	ContactInfo ContactInfo

	XXXX string `fuzz:"ignore"`
}

type Education struct {
	School string `fuzz:"length(8,16)"`
	From   string `fuzz:"date(2006-01-02)"`
	To     string `fuzz:"date(2006-01-02)"`
}

type ContactInfo struct {
	Mobile string `fuzz:"length(7,11)"`
	Home   string `fuzz:"length(7,11)"`
	Work   string `fuzz:"length(7,11)"`
	Email  string `fuzz:"length(10,64)"`
}
