package entity

type PacketState string

const (
	STREAMING   PacketState = "STREAMING"
	FINISHED    PacketState = "FINISHED"
	REQUESTING  PacketState = "REQUESTING"
	ACKNOWLEDGE PacketState = "ACKNOWLEDGE"
)

type Database struct {
	Videos map[string][]byte
}

type Packet struct {
	Id    []byte
	Data  []byte
	State []byte
	//Timestamp time.Time //TODO: Add Timestamp to Packet
}

type Node struct {
	Address       string `json:"address"`
	Port          string `json:"port"`
	FullAddress   string `json:"fullAddress"`
	Type          string `json:"type"`
	BootstrapPort string `json:"bootstrapPort"`
	Neighbours    []Node `json:"neighbours"`
}
