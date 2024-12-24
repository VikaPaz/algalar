package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/VikaPaz/algalar/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Repository struct {
	conn *sql.DB
	log  *logrus.Logger
}

func NewRepository(conn *sql.DB, logger *logrus.Logger) *Repository {
	return &Repository{
		conn: conn,
		log:  logger,
	}
}

func (r *Repository) CreateUser(user models.User) (string, error) {
	query := `
        INSERT INTO users (inn, name, surname, gender, login, password, utc_timezone, phone)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id`

	var userID string
	err := r.conn.QueryRow(query, user.INN, user.Name, user.Surname, user.Gender, user.Login, user.Password, user.Timezone, user.Phone).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (r *Repository) GetById(userID string) (models.User, error) {
	query := `
        SELECT inn, name, surname, gender, login, password, utc_timezone, phone
        FROM users
        WHERE id = $1`

	user := models.User{}
	err := r.conn.QueryRow(query, userID).Scan(&user.INN, &user.Name, &user.Surname, &user.Gender, &user.Login, &user.Password, &user.Timezone, &user.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, models.ErrNoContent
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *Repository) ChangePassword(userID, newPassword string) error {
	if newPassword == "" {
		return errors.New("new password is required")
	}

	query := `
        UPDATE users
        SET password = $1
        WHERE id = $2`

	_, err := r.conn.Exec(query, newPassword, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetIDByLoginAndPassword(email, password string) (string, error) {
	if email == "" || password == "" {
		return "", errors.New("email and password are required")
	}

	query := `
        SELECT id
        FROM  users
        WHERE login = $1 AND password = $2`

	var userID string
	err := r.conn.QueryRow(query, email, password).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", models.ErrNoContent
		}
		return "", err
	}

	return userID, nil
}

// Auto
func (r *Repository) CreateCar(car models.Car) (models.Car, error) {
	query := `
        INSERT INTO cars (id_company, state_number, brand, device_number, id_unicum, count_axis, car_type)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING *`
	resp := models.Car{}
	err := r.conn.QueryRow(query, car.IDCompany, car.StateNumber, car.Brand, car.DeviceNumber, car.IDUnicum, car.CountAxis, car.Type).Scan(
		&resp.ID,
		&resp.IDCompany,
		&resp.StateNumber,
		&resp.Brand,
		&resp.DeviceNumber,
		&resp.IDUnicum,
		&resp.Type,
		&resp.CountAxis,
	)
	if err != nil {
		return models.Car{}, err
	}

	return resp, nil
}

func (r *Repository) GetCarById(carID string) (models.Car, error) {
	query := `
		SELECT id, id_company, state_number, brand, device_number, id_unicum, count_axis
		FROM cars
		WHERE id = $1`

	car := models.Car{}
	err := r.conn.QueryRow(query, carID).Scan(
		&car.ID,
		&car.IDCompany,
		&car.StateNumber,
		&car.Brand,
		&car.DeviceNumber,
		&car.IDUnicum,
		&car.CountAxis,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Car{}, models.ErrNoContent
		}
		return models.Car{}, err
	}

	return car, nil
}

func (r *Repository) GetIdCarByStateNumber(stateNumber string) (string, error) {
	query := `
		SELECT id
		FROM cars
		WHERE state_number = $1`

	var carID string
	err := r.conn.QueryRow(query, stateNumber).Scan(&carID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", models.ErrNoContent
		}
		return "", err
	}

	return carID, nil
}

func (r *Repository) GetCarsList(userID string, offset int, limit int) ([]models.Car, error) {
	query := `
		SELECT id, id_company, state_number, brand, device_number, id_unicum, count_axis
		FROM cars
		WHERE id_company = $1
		LIMIT $2 OFFSET $3`

	cars := []models.Car{}

	rows, err := r.conn.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var car models.Car
		if err := rows.Scan(&car.ID, &car.IDCompany, &car.StateNumber, &car.Brand, &car.DeviceNumber, &car.IDUnicum, &car.CountAxis); err != nil {
			return nil, err
		}
		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}

func (r *Repository) GetCarByStateNumber(stateNumber string) (models.Car, error) {
	query := `
	SELECT id, id_company, state_number, brand, device_number, id_unicum, count_axis
	FROM cars
	WHERE state_number = $1`

	car := models.Car{}
	err := r.conn.QueryRow(query, stateNumber).Scan(
		&car.ID,
		&car.IDCompany,
		&car.StateNumber,
		&car.Brand,
		&car.DeviceNumber,
		&car.IDUnicum,
		&car.CountAxis,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Car{}, models.ErrNoContent
		}
		return models.Car{}, err
	}

	return car, nil
}

// Wheel
func (r *Repository) CreateWheel(wheel models.Wheel) (string, error) {
	query := `
        INSERT INTO wheels (id_company, id_car, count_axis, position, sensor_number, size, cost, brand, model, mileage, min_temperature, min_pressure, max_temperature, max_pressure, ngp, tkvh)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        RETURNING id`

	var wheelID string
	err := r.conn.QueryRow(query, wheel.IDCompany, wheel.IDCar, wheel.AxisNumber, wheel.Position, wheel.SensorNumber, wheel.Size, wheel.Cost, wheel.Brand, wheel.Model, wheel.Mileage, wheel.MinTemperature, wheel.MinPressure, wheel.MaxTemperature, wheel.MaxPressure, *wheel.Ngp, *wheel.Tkvh).Scan(&wheelID)
	if err != nil {
		return "", err
	}

	return wheelID, nil
}

func (r *Repository) GetWheelsByStateNumber(stateNumber string) ([]models.Wheel, error) {
	query := `
    SELECT
        w.id AS wheel_id,
        w.id_company,
        w.id_car,
        w.count_axis AS axis_number,
        w.position,
        w.size,
        w.cost,
        w.brand,
        w.model,
        w.ngp,
        w.tkvh,
        w.mileage,
        w.min_temperature,
        w.min_pressure,
        w.max_temperature,
        w.max_pressure
    FROM
        wheels w
    JOIN
        cars c ON w.id_car = c.id
    WHERE
        c.state_number = $1
`

	var wheels []models.Wheel

	rows, err := r.conn.Query(query, stateNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var wheel models.Wheel
		err := rows.Scan(
			&wheel.ID,
			&wheel.IDCompany,
			&wheel.IDCar,
			&wheel.AxisNumber,
			&wheel.Position,
			&wheel.Size,
			&wheel.Cost,
			&wheel.Brand,
			&wheel.Model,
			&wheel.Ngp,
			&wheel.Tkvh,
			&wheel.Mileage,
			&wheel.MinTemperature,
			&wheel.MinPressure,
			&wheel.MaxTemperature,
			&wheel.MaxPressure,
		)

		if err != nil {
			r.log.Errorf("scan faild: %v", err)
			return nil, err
		}

		r.log.Debugf("wheel: %v", wheel)

		wheels = append(wheels, wheel)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(wheels) == 0 {
		return nil, models.ErrNoContent
	}

	r.log.Debugf("query resp: %v", wheels)

	return wheels, nil
}

func (r *Repository) GetWheelById(wheelID string) (models.Wheel, error) {
	query := `
        SELECT id_car, count_axis, position, size, cost, brand, model, mileage, min_temperature, min_pressure, max_temperature, max_pressure
        FROM wheels
        WHERE id = $1`

	wheel := models.Wheel{}
	err := r.conn.QueryRow(query, wheelID).Scan(&wheel.IDCar, &wheel.AxisNumber, &wheel.Position, &wheel.Size, &wheel.Cost, &wheel.Brand, &wheel.Model, &wheel.Mileage, &wheel.MinTemperature, &wheel.MinPressure, &wheel.MaxTemperature, &wheel.MaxPressure)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Wheel{}, models.ErrNoContent
		}
		return models.Wheel{}, err
	}

	return wheel, nil
}

func (r *Repository) CreateBreakage(breakage models.Breakage) (string, error) {
	query := `
        INSERT INTO breakages (car_id, state_number, type, description, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	var breakageID string
	err := r.conn.QueryRow(query,
		breakage.CarID, breakage.StateNumber, breakage.Type, breakage.Description, breakage.Datetime).
		Scan(&breakageID)
	if err != nil {
		return "", err
	}

	return breakageID, nil
}

func (r *Repository) ChangeWheel(wheel models.Wheel) error {
	carID, err := uuid.Parse(wheel.IDCar)
	if err != nil {
		return fmt.Errorf("error parsing carID '%s' into UUID: %w", carID, err)
	}

	query := `
        UPDATE wheels
        SET id_car = $1, count_axis = $2, position = $3, size = $4, cost = $5, brand = $6, model = $7, mileage = $8, min_temperature = $9, min_pressure = $10, max_temperature = $11, max_pressure = $12, ngp = $13, tkvh = $14
        WHERE id_car = $15 AND position = $16`
	err = r.conn.QueryRow(query, carID, wheel.AxisNumber, wheel.Position, wheel.Size, wheel.Cost, wheel.Brand, wheel.Model, wheel.Mileage, wheel.MinTemperature, wheel.MinPressure, wheel.MaxTemperature, wheel.MaxPressure, *wheel.Ngp, *wheel.Tkvh, carID, wheel.Position).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetBreakagesByCarId(carID string) ([]models.Breakage, error) {
	query := `
        SELECT id, car_id, state_number, type, description, created_at
        FROM breakages
        WHERE car_id = $1
    `

	var breakages []models.Breakage

	parsedUUID, err := uuid.Parse(carID)
	if err != nil {
		return nil, fmt.Errorf("error parsing carID '%s' into UUID: %w", carID, err)
	}

	rows, err := r.conn.Query(query, parsedUUID)
	if err != nil {
		return nil, fmt.Errorf("error executing query to get breakages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var breakage models.Breakage
		if err := rows.Scan(&breakage.ID, &breakage.CarID, &breakage.StateNumber, &breakage.Type, &breakage.Description, &breakage.Datetime); err != nil {
			return nil, fmt.Errorf("error scanning row into breakage: %w", err)
		}
		breakages = append(breakages, breakage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return breakages, nil
}

func (r *Repository) SelectAny(table string, key string, val any) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1 LIMIT 1", table, key)

	var exists int
	err := r.conn.QueryRow(query, val).Scan(&exists)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error querying database: %v", err)
	}

	return exists == 1, nil
}

// Sensors
func (r *Repository) CreateData(newData models.SensorData) (models.SensorData, error) {
	query := `INSERT INTO sensors_data (device_number, sensor_number, pressure, temperature) 
	VALUES ($1, $2, $3, $4) 
	RETURNING id, device_number, sensor_number, pressure, temperature`

	var result models.SensorData
	err := r.conn.QueryRow(query, newData.DeviceNumber, newData.SensorNumber, newData.Pressure, newData.Temperature).
		Scan(&result.ID, &result.DeviceNumber, &result.SensorNumber, &result.Pressure, &result.Temperature)
	if err != nil {
		return models.SensorData{}, err
	}

	return result, nil
}

func (r *Repository) SensorsDataByCarID(carID string) ([]models.SensorsData, error) {
	query := `SELECT s.id, s.device_number, s.sensor_number, w.wheel_position, s.pressure, s.temperature 
	FROM sensors_data s 
	JOIN cars c ON s.device_number = c.device_number 
	JOIN wheels w ON s.id = w.id_car 
	WHERE c.id = $1`

	rows, err := r.conn.Query(query, carID)
	if err != nil {
		return []models.SensorsData{}, err
	}
	defer rows.Close()

	var sensorsData []models.SensorsData
	for rows.Next() {
		var id uuid.UUID
		var deviceNumber string
		var sensorNumber string
		var wheelPosition int
		var pressure float32
		var temperature float32

		err := rows.Scan(&id, &deviceNumber, &sensorNumber, &wheelPosition, &pressure, &temperature)
		if err != nil {
			return []models.SensorsData{}, err
		}

		sensorsData = append(sensorsData, models.SensorsData{
			WheelPosition: wheelPosition,
			Pressure:      pressure,
			Temperature:   temperature,
		})
	}
	return sensorsData, nil
}

// Data
func (r *Repository) Temperaturedata(filter models.TemperatureDataByWheelIDFilter) ([]models.TemperatureData, error) {
	query := `SELECT s.temperature, s.created_at 
	FROM sensors_data s 
	JOIN wheels w ON s.sensor_number = w.sensor_number 
	WHERE w.id = $1 AND s.created_at 
	BETWEEN $2 AND $3`

	rows, err := r.conn.Query(query, filter.IDWheel, filter.From, filter.To)
	if err != nil {
		return []models.TemperatureData{}, err
	}
	defer rows.Close()

	var temperatureData []models.TemperatureData
	for rows.Next() {
		var temp models.TemperatureData

		err := rows.Scan(&temp.Temperature, &temp.Datetime)
		if err != nil {
			return []models.TemperatureData{}, err
		}

		temperatureData = append(temperatureData, temp)
	}

	return temperatureData, nil
}

func (r *Repository) Pressuredata(filter models.PressureDataByWheelIDFilter) ([]models.PressureData, error) {
	query := `SELECT s.pressure, s.created_at 
	FROM sensors_data s 
	JOIN wheels w ON s.sensor_number = w.sensor_number 
	WHERE w.id = $1 AND s.created_at BETWEEN $2 AND $3`

	rows, err := r.conn.Query(query, filter.IDWheel, filter.From, filter.To)
	if err != nil {
		return []models.PressureData{}, err
	}
	defer rows.Close()

	var pressureData []models.PressureData
	for rows.Next() {
		var press models.PressureData

		err := rows.Scan(&press.Pressure, &press.Datetime)
		if err != nil {
			return []models.PressureData{}, err
		}

		pressureData = append(pressureData, press)
	}

	return pressureData, nil
}

// Driver
func (r *Repository) CreateDriver(driver models.Driver) (models.Driver, error) {
	query := `
	INSERT INTO drivers (id_company, id_car, name, surname, middle_name, phone, birthday, rating, worked_time)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, id_company, id_car, name, surname, middle_name, phone, birthday, rating, worked_time, created_at
	`

	resp := models.Driver{}
	err := r.conn.QueryRow(query, driver.IDCompany, driver.IDCar, driver.Name, driver.Surname, driver.Middle, driver.Phone, driver.Birthday, driver.Rating, driver.WorkedTime).Scan(
		&resp.ID,
		&resp.IDCompany,
		&resp.IDCar,
		&resp.Name,
		&resp.Surname,
		&resp.Middle,
		&resp.Phone,
		&resp.Birthday,
		&resp.Rating,
		&resp.WorkedTime,
		&resp.CreatedAt,
	)
	if err != nil {
		return models.Driver{}, fmt.Errorf("failed to create driver or find car: %w", err)
	}

	return resp, nil
}

func (r *Repository) GetDriversList(userID string, limit int, offset int) ([]models.DriverStatisticsResponse, error) {
	query := `
	SELECT 
		CONCAT(d.name, ' ', d.surname, ' ', COALESCE(d.middle_name, '')) AS full_name,
		d.worked_time,
		(EXTRACT(YEAR FROM AGE(d.birthday))) AS experience,
		d.rating,
		COALESCE(COUNT(b.id), 0) AS breakages_count,
		d.id AS driver_id
	FROM drivers d
	LEFT JOIN breakages b ON d.id_car = b.id_car
	WHERE d.id_company = $1
	GROUP BY d.id
	ORDER BY d.created_at DESC
	LIMIT $2 OFFSET $3
	`

	rows, err := r.conn.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get drivers list: %w", err)
	}
	defer rows.Close()

	var drivers []models.DriverStatisticsResponse

	for rows.Next() {
		var driver models.DriverStatisticsResponse
		err := rows.Scan(
			&driver.FullName,
			&driver.WorkedTime,
			&driver.Experience,
			&driver.Rating,
			&driver.BreakagesCount,
			&driver.DriverID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan driver data: %w", err)
		}

		drivers = append(drivers, driver)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return drivers, nil
}

// Report
func (r *Repository) GetReportData(userId string) ([]models.ReportData, error) {
	query := `
		SELECT
			w.id AS wheel_id,             
			c.state_number,              
			w.brand AS tire_brand,       
			w.mileage,                   
			COUNT(CASE WHEN s.temperature < w.min_temperature OR s.temperature > w.max_temperature THEN 1 END) AS temp_out_of_bounds,  
			COUNT(CASE WHEN s.pressure < w.min_pressure OR s.pressure > w.max_pressure THEN 1 END) AS pressure_out_of_bounds  
		FROM
			cars c
		JOIN wheels w ON w.id_car = c.id
		JOIN sensors_data s ON s.device_number = c.device_number AND s.sensor_number = w.sensor_number
		WHERE
			c.id_company = $1  
		GROUP BY
			w.id, c.state_number, w.brand, w.mileage, w.position  
		ORDER BY
			c.state_number, w.position; 
	`

	rows, err := r.conn.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var reportData []models.ReportData

	for rows.Next() {
		var data models.ReportData
		if err := rows.Scan(&data.IdWheel, &data.StateNumber, &data.TireBrand, &data.Mileage, &data.TempOutOfBounds, &data.PressureOutOfBounds); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		reportData = append(reportData, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return reportData, nil
}

func (r *Repository) GetCarWheelData(carID string) (models.CarWithWheels, error) {
	query := `
        SELECT
            c.id AS car_id,
            c.state_number,
            c.brand,
            c.device_number,
            c.id_unicum,
            c.count_axis,
			c.car_type,
            w.id AS wheel_id,
            w.count_axis AS wheel_count_axis,
            w.position AS wheel_position,
            w.size AS wheel_size,
            w.cost AS wheel_cost,
            w.brand AS wheel_brand,
            w.model AS wheel_model,
            w.ngp AS wheel_ngp,
            w.tkvh AS wheel_tkvh,
            w.mileage AS wheel_mileage,
            w.min_temperature AS wheel_min_temperature,
            w.min_pressure AS wheel_min_pressure,
            w.max_temperature AS wheel_max_temperature,
            w.max_pressure AS wheel_max_pressure
        FROM
            cars c
        LEFT JOIN
            wheels w ON c.id = w.id_car
        WHERE
            c.id = $1;`

	var wheels []models.Wheel
	var car models.CarWithWheels

	rows, err := r.conn.Query(query, carID)
	if err != nil {
		return car, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var wheel models.Wheel
		var carID, wheelID uuid.UUID
		var position, countAxis int
		var size, cost, ngp, tkvh, mileage, minTemp, minPressure, maxTemp, maxPressure float32
		var brand, model string

		err := rows.Scan(
			&carID,
			&car.StateNumber,
			&car.Brand,
			&car.DeviceNumber,
			&car.IDUnicum,
			&car.CountAxis,
			&car.AutoType,
			&wheelID,
			&countAxis,
			&position,
			&size,
			&cost,
			&brand,
			&model,
			&ngp,
			&tkvh,
			&mileage,
			&minTemp,
			&minPressure,
			&maxTemp,
			&maxPressure,
		)
		if err != nil {
			return car, fmt.Errorf("error scanning row: %w", err)
		}

		wheel.ID = wheelID.String()
		wheel.AxisNumber = countAxis
		wheel.Position = position
		wheel.Size = size
		wheel.Cost = cost
		wheel.Brand = brand
		wheel.Model = model
		wheel.Ngp = &ngp
		wheel.Tkvh = &tkvh
		wheel.Mileage = mileage
		wheel.MinTemperature = minTemp
		wheel.MinPressure = minPressure
		wheel.MaxTemperature = maxTemp
		wheel.MaxPressure = maxPressure

		wheels = append(wheels, wheel)
	}

	if err := rows.Err(); err != nil {
		return car, fmt.Errorf("error after scanning rows: %w", err)
	}

	car.ID = carID
	car.Wheels = wheels

	return car, nil
}
