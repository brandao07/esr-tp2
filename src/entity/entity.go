package entity

type PacketState string

const (
	STREAMING   PacketState = "STREAMING"
	FINISHED    PacketState = "FINISHED"
	REQUESTING  PacketState = "REQUESTING"
	ACKNOWLEDGE PacketState = "ACKNOWLEDGE"
)

type Database struct {
	Data map[string]string
}

type Packet struct {
	Id    []byte
	Data  []byte
	State []byte
}

type Node struct {
	ID         int    `json:"id"`
	Label      string `json:"label"`
	Address    string `json:"address"`
	Type       string `json:"type"`
	Neighbours []Node `json:"neighbours"`
}
