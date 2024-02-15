package node

type Edge struct {
	SrcID, SrcName string
	DstID, DstName string
	Port           *Port
}
