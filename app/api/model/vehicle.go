package model

type Vehicle struct {
	ID             string `json:"id"`              //用户ID
	TransmitterId  string `json:"transmitter_id"`  //车辆id
	ReceiverId     string `json:"receiver_id"`     //发射机id
	VehicleBattery string `json:"vehicle_battery"` //车辆电量
	VehicleState   int32  `json:"vehicle_state"`   //在线状态
}

func (model Vehicle) GetID() any {
	return model.ID
}

func (model Vehicle) TableName() string {
	return "vehicle"
}
