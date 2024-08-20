package files

type MockLoader struct {
}

func (loader MockLoader) Load(path string) (string, error) {
	return `Id,Date,Transaction
	1,8/19,+3.23
	2,9/22,+11.11
	3,8/11,-1.10
	4,9/1,+0.23
	`, nil
}

var _ Loader = SystemLoader{}
