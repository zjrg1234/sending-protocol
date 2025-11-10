package repo

import (
	"megin/app/api/model"
	"megin/system/datasource"
)

func GetVehicleConfig(ReceiverId interface{}) (*model.ReceiverTransmitterConfig, error) {
	db := datasource.GetDB()
	rows := &model.ReceiverTransmitterConfig{}

	err := db.First(&rows, "receiver_id = ?", ReceiverId).Error

	return rows, err
}

func UpdateVehicleConfig(dataList *model.ReceiverTransmitterConfig) error {
	db := datasource.GetDB()

	err := db.Model(&model.ReceiverTransmitterConfig{}).Where("id = ?", dataList.ID).Update("receiver_host_port", dataList.ReceiverHostPort).Error
	return err
}

func GetVehicle(ReceiverId interface{}) (*model.Vehicle, error) {
	db := datasource.GetDB()
	rows := &model.Vehicle{}
	err := db.First(&rows, "receiver_id = ?", ReceiverId).Error

	return rows, err
}

func UpdateVehicle(dataList *model.Vehicle) error {
	db := datasource.GetDB()

	err := db.Model(&model.Vehicle{}).Where("id = ?", dataList.ID).Update("vehicle_battery", dataList.VehicleBattery).Error
	return err
}
