package namegenerator

import petname "github.com/dustinkirkland/golang-petname"

type NameGenerator struct {
	numWords  int
	separator string
}

func New(numWords int, separator string) *NameGenerator {
	petname.NonDeterministicMode()
	return &NameGenerator{
		numWords:  numWords,
		separator: separator,
	}
}

func (n *NameGenerator) Generate() string {
	return petname.Generate(n.numWords, n.separator)
}
