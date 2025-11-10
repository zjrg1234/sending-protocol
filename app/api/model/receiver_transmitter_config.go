package model

type ReceiverTransmitterConfig struct {
	ID                  string `json:"id"` //用户ID
	VehicleId           string `json:"vehicle_id"`
	TransmitterId       string `json:"transmitter_id"`
	ReceiverId          string `json:"receiver_id"`
	ReceiverHostPort    string `json:"receiver_host_port"`
	TransmitterHostPort string `json:"transmitter_host_port"`
}

//type UpdateReceiverTransmitterConfig struct {
//	ID               string `json:"id"` //用户ID
//	ReceiverHostPort string `json:"receiver_host_port"`
//}

func (model ReceiverTransmitterConfig) GetID() any {
	return model.ID
}

func (model ReceiverTransmitterConfig) TableName() string {
	return "receiver_transmitter_config"
}
