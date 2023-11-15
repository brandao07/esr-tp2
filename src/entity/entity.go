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
	Address       string `json:"address"`
	Port          string `json:"port"`
	FullAddress   string `json:"fullAddress"`
	Type          string `json:"type"`
	BootstrapPort string `json:"bootstrapPort"`
	Neighbours    []Node `json:"neighbours"`
}
