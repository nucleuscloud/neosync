package dbconnectconfig

type DbConnectConfig interface {
	String() string
	GetUser() string
}
