package namegenerator

import petname "github.com/dustinkirkland/golang-petname"

type NameGenerator struct{}

func New() *NameGenerator {
	petname.NonDeterministicMode()
	return &NameGenerator{}
}

func (n *NameGenerator) Generate() string {
	return petname.Generate(2, "-")
}
