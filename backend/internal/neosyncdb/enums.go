package neosyncdb

type AccountType int16

const (
	AccountType_Personal AccountType = iota
	AccountType_Team
	AccountType_Enterprise
)
