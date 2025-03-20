package tablesync_shared

const (
	AllocateIdentityBlock = "allocate-identity-block"
)

type AllocateIdentityBlockRequest struct {
	Id        string // This will be the value present in the ColumnIdentityCursors map (key)
	BlockSize uint   // The size of the block the caller wishes to allocate
}
type AllocateIdentityBlockResponse struct {
	StartValue uint // Inclusive
	EndValue   uint // Exclusive - represents the next value after the last valid value
}
