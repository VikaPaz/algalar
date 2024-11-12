package models

import "time"

type User struct {
	ID         string
	INN        int
	Name       string
	Surname    string
	MiddleName string
	Login      string
	Password   string
	Timezone   string
}

type Car struct {
	ID          string
	IDCompany   string
	StateNumber string
	Brand       string
	IDDevice    string
	IDUnicum    string
	CountAxis   int
}

type Wheel struct {
	ID             string
	IDCompany      string
	IDCar          string
	AxisNumber     int
	Position       int
	Size           float32
	Cost           float32
	Brand          string
	Model          string
	Mileage        float32
	MinTemperature float32
	MinPressure    float32
	MaxTemperature float32
	MaxPressure    float32
}

type Sensor struct {
	ID          string
	CarID       string
	StateNumber string
	CountAxis   int
	Position    int
	Pressure    float32
	Temperature float32
	Datetime    time.Time
}

type Breakage struct {
	ID          string
	CarID       string
	StateNumber string
	Type        string
	Description string
	Datetime    time.Time
}

type GetReportParams struct {
	UserId string `form:"userId" json:"userId"`
}

type GetSensorParams struct {
	WheelId string `form:"wheelId" json:"wheelId"`
}
