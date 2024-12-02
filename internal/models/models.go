package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type User struct {
	ID       string
	INN      string
	Name     string
	Surname  string
	Gender   string
	Login    string
	Password string
	Timezone string
	Phone    string
}

type Car struct {
	ID          string
	IDCompany   string
	StateNumber string
	Brand       string
	IDDevice    string
	IDUnicum    string
	CountAxis   int
	Type        string
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
	Ngp            *float32
	Tkvh           *float32
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

type ReportData struct {
	IdWheel             string
	StateNumber         string
	TireBrand           string
	Mileage             float32
	TempOutOfBounds     int
	PressureOutOfBounds int
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type CarWithWheels struct {
	ID          uuid.UUID `json:"id"`
	StateNumber string    `json:"state_number"`
	Brand       string    `json:"brand"`
	IDDevice    string    `json:"id_device"`
	IDUnicum    string    `json:"id_unicum"`
	CountAxis   int       `json:"count_axis"`
	Wheels      []Wheel   `json:"wheels"`
}
