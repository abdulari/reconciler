package parser

type Step struct {
	ID           string
	Name         string
	VerifyScript string
	ExecScript   string
}

type Parser interface {
	Parse() ([]Step, error)
}