package server

import "net/http"

// GetReportParams defines parameters for GetReport.
type GetReportParams struct {
	UserId string `form:"userId" json:"userId"`
}

// GetSensorParams defines parameters for GetSensor.
type GetSensorParams struct {
	WheelId string `form:"wheelId" json:"wheelId"`
}

// Server implementation
type ServImplemented struct{}

// Register a vehicle
// (POST /auto)
func (_ ServImplemented) PostAuto(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// User login
// (POST /login)
func (_ ServImplemented) PostLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Refresh access token
// (POST /refresh)
func (_ ServImplemented) PostRefresh(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Generate report
// (GET /report)
func (_ ServImplemented) GetReport(w http.ResponseWriter, r *http.Request, params GetReportParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get sensor data
// (GET /sensor)
func (_ ServImplemented) GetSensor(w http.ResponseWriter, r *http.Request, params GetSensorParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// User registration
// (POST /user)
func (_ ServImplemented) PostUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Update user password
// (PUT /user)
func (_ ServImplemented) PutUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get user details
// (GET /user/{id})
func (_ ServImplemented) GetUserId(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Register a wheel
// (POST /wheels)
func (_ ServImplemented) PostWheels(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Update wheel data
// (PUT /wheels)
func (_ ServImplemented) PutWheels(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Get wheel data
// (GET /wheels/{id})
func (_ ServImplemented) GetWheelsId(w http.ResponseWriter, r *http.Request, id string) {
	w.WriteHeader(http.StatusNotImplemented)
}
