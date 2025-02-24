package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// User
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

// UpdateUser updates user information in the database and returns the updated user ID.
func (r *Repository) UpdateUser(user models.User) (string, error) {
	query := `
        UPDATE users 
        SET inn = $1, 
            name = $2, 
            surname = $3, 
            gender = $4, 
            login = $5, 
            utc_timezone = $6, 
            phone = $7
        WHERE id = $8
        RETURNING id`

	r.log.Debugf("Executing query to update user with ID: %s", user.ID)

	var userID string
	err := r.conn.QueryRow(query,
		user.INN, user.Name, user.Surname, user.Gender,
		user.Login, user.Timezone, user.Phone, user.ID,
	).Scan(&userID)

	if err != nil {
		r.log.Errorf("Failed to update user: %v", err)
		return "", fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("User updated successfully: %s", userID)
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

func (r *Repository) GetCarByDeviceNumber(ctx context.Context, device string) (models.Car, error) {
	query := `
		SELECT id, id_company, state_number, brand, device_number, id_unicum, car_type, count_axis 
		FROM cars 
		WHERE device_number = $1
	`

	r.log.Debugf("Executing query: %s with value: %s", query, device)

	var car models.Car
	err := r.conn.QueryRowContext(ctx, query, device).Scan(
		&car.ID,
		&car.IDCompany,
		&car.StateNumber,
		&car.Brand,
		&car.DeviceNumber,
		&car.IDUnicum,
		&car.Type,
		&car.CountAxis,
	)

	if err != nil {
		r.log.Errorf("Failed to get car by device number: %v", err)
		return models.Car{}, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("Car retrieved successfully: %+v", car)
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

	r.log.Debugf("Executing query to fetch car list: userID=%s, limit=%d, offset=%d", userID, limit, offset)

	cars := []models.Car{}
	rows, err := r.conn.Query(query, userID, limit, offset)
	if err != nil {
		r.log.Errorf("%v: %v", models.ErrFailedToExecuteQuery, err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}
	defer rows.Close()

	for rows.Next() {
		var car models.Car
		if err := rows.Scan(
			&car.ID,
			&car.IDCompany,
			&car.StateNumber,
			&car.Brand,
			&car.DeviceNumber,
			&car.IDUnicum,
			&car.CountAxis,
		); err != nil {
			r.log.Errorf("%v: %v", models.ErrFailedToScanRow, err)
			return nil, fmt.Errorf("%w: %v", models.ErrFailedToScanRow, err)
		}
		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorf("%v: %v", models.ErrFailedToIterateRows, err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToIterateRows, err)
	}

	if len(cars) == 0 {
		r.log.Debugf("No cars found for userID=%s", userID)
		return nil, models.ErrNoContent
	}

	r.log.Debugf("Successfully fetched %d cars for userID=%s", len(cars), userID)
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
		w.sensor_number,
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
			&wheel.SensorNumber,
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

func (r *Repository) GetBreakagesByCarId(carID string) ([]models.BreakageInfo, error) {
	query := `
        SELECT 
			b.id, 
			CONCAT(d.name, ' ', d.surname, ' ', COALESCE(d.middle_name, '')) AS full_name,
			c.state_number, 
			b.type, 
			b.description, 
			b.created_at
        FROM breakages b
        JOIN cars c ON b.id_car = c.id
		JOIN drivers d ON d.id_car = c.id
        WHERE b.id_car = $1
		;
    `

	var breakages []models.BreakageInfo

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
		var breakage models.BreakageInfo
		if err := rows.Scan(
			&breakage.ID,
			&breakage.DriverName,
			&breakage.StateNumber,
			&breakage.Type,
			&breakage.Description,
			&breakage.CreatedAt,
		); err != nil {
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
	query := `INSERT INTO sensors_data (device_number, sensor_number, pressure, temperature, created_at) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id, device_number, sensor_number, pressure, temperature, created_at`

	var result models.SensorData
	err := r.conn.QueryRow(query, newData.DeviceNumber, newData.SensorNumber, newData.Pressure, newData.Temperature, newData.Time).
		Scan(&result.ID, &result.DeviceNumber, &result.SensorNumber, &result.Pressure, &result.Temperature, &result.Time)
	if err != nil {
		return models.SensorData{}, err
	}

	return result, nil
}

func (r *Repository) SensorsDataByCarID(carID string) ([]models.SensorsData, error) {
	r.log.Debugf("Querying for sensors data with carID: %v", carID)

	query := `WITH latest_data AS (
		SELECT 
			s.id,
			s.device_number,
			s.sensor_number,
			w.position,
			s.pressure,
			s.temperature,
			ROW_NUMBER() OVER (PARTITION BY w.position ORDER BY s.created_at DESC) AS rn
			FROM sensors_data s
		JOIN cars c ON s.device_number = c.device_number
		JOIN wheels w ON s.sensor_number = w.sensor_number
		WHERE c.id = $1
		)
	SELECT id, device_number, sensor_number, position, pressure, temperature
	FROM latest_data
	WHERE rn = 1`

	r.log.Debugf("Executing query: %v", query)

	rows, err := r.conn.Query(query, carID)
	if err != nil {
		r.log.Errorf("Error executing query: %v", err)
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
			r.log.Errorf("Error scanning row: %v", err)
			return []models.SensorsData{}, err
		}

		r.log.Debugf("Fetched data for wheel position %v: Pressure = %v, Temperature = %v", wheelPosition, pressure, temperature)

		sensorsData = append(sensorsData, models.SensorsData{
			WheelPosition: wheelPosition,
			Pressure:      pressure,
			Temperature:   temperature,
		})
	}

	r.log.Debugf("Fetched %d sensor data entries for carID: %v", len(sensorsData), carID)

	return sensorsData, nil
}

// Data
func (r *Repository) Temperaturedata(filter models.TemperatureDataByWheelIDFilter) ([]models.TemperatureData, error) {
	query := `SELECT s.temperature, s.created_at 
	FROM sensors_data s 
	JOIN wheels w ON s.sensor_number = w.sensor_number 
	WHERE w.id = $1 AND s.created_at 
	BETWEEN $2 AND $3
	ORDER BY s.created_at`

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
	WHERE w.id = $1 AND s.created_at BETWEEN $2 AND $3
	ORDER BY s.created_at`

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
	EXTRACT(YEAR FROM AGE(d.created_at)) * 12 + EXTRACT(MONTH FROM AGE(d.created_at)) AS experience_months,
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

// func (r *Repository) GetDriverInfo(driverID string) (models.DriverStatisticsResponse, error) {
// 	query := `
// 	SELECT
// 		CONCAT(d.name, ' ', d.surname, ' ', COALESCE(d.middle_name, '')) AS full_name,
// 		d.worked_time,
// 		EXTRACT(YEAR FROM AGE(d.created_at)) * 12 + EXTRACT(MONTH FROM AGE(d.created_at)) AS experience_months,
// 		d.rating,
// 		COALESCE(COUNT(b.id), 0) AS breakages_count,
// 		d.id AS driver_id
// 	FROM drivers d
// 	LEFT JOIN breakages b ON d.id_car = b.id_car
// 	WHERE d.id = $1
// 	GROUP BY d.id
// 	`

// 	rows := r.conn.QueryRow(query, driverID)

// 	var driver models.DriverStatisticsResponse
// 	err := rows.Scan(
// 		&driver.FullName,
// 		&driver.WorkedTime,
// 		&driver.Experience,
// 		&driver.Rating,
// 		&driver.BreakagesCount,
// 		&driver.DriverID,
// 	)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return models.DriverStatisticsResponse{}, models.ErrDriverNotFound
// 		}
// 		return models.DriverStatisticsResponse{}, fmt.Errorf("failed to fetch driver info: %w", err)
// 	}

// 	return driver, nil
// }

func (r *Repository) GetDriverInfo(driverID string) (models.DriverInfoResponse, error) {
	query := `
		SELECT name, surname, middle_name, phone, birthday
		FROM drivers
		WHERE id = $1
	`

	var driverInfo models.DriverInfoResponse
	err := r.conn.QueryRow(query, driverID).Scan(
		&driverInfo.Name,
		&driverInfo.Surname,
		&driverInfo.MiddleName,
		&driverInfo.Phone,
		&driverInfo.Birthday,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.DriverInfoResponse{}, models.ErrDriverNotFound
		}
		return models.DriverInfoResponse{}, fmt.Errorf("failed to fetch driver info: %w", err)
	}

	return driverInfo, nil
}

func (r *Repository) GetDriverByCaDviceNum(ctx context.Context, deviceNum string) (models.Driver, error) {
	query := `
		WITH car_info AS (
			SELECT id
			FROM cars 
			WHERE device_number = $1
			LIMIT 1
		)
		SELECT 
			id, id_company, id_car, name, surname, middle_name, phone, birthday, rating, worked_time, created_at
			FROM drivers
		WHERE id_car = (SELECT id FROM car_info)
		;
	`

	var driverInfo models.Driver
	err := r.conn.QueryRow(query, deviceNum).Scan(
		&driverInfo.ID,
		&driverInfo.IDCompany,
		&driverInfo.IDCar,
		&driverInfo.Name,
		&driverInfo.Surname,
		&driverInfo.Middle,
		&driverInfo.Phone,
		&driverInfo.Birthday,
		&driverInfo.Rating,
		&driverInfo.WorkedTime,
		&driverInfo.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Driver{}, models.ErrDriverNotFound
		}
		return models.Driver{}, fmt.Errorf("failed to fetch driver info: %w", err)
	}

	return driverInfo, nil
}

func (r *Repository) UpdateDriverWorktime(deviceNum string, workedTime int) error {
	query := `
		UPDATE drivers
		SET worked_time = worked_time + $1
		WHERE id_car = (SELECT id FROM cars WHERE device_number = $2)
		`

	res, err := r.conn.Exec(query, workedTime, deviceNum)
	if err != nil {
		return fmt.Errorf("failed to update driver worktime: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrDriverNotFound
	}

	return nil
}

// Position
// CreatePosition creates a new position entry in the database and returns the created position.
func (r *Repository) CreatePosition(ctx context.Context, position models.Position) (models.Position, error) {
	query := `
		INSERT INTO position_data (device_number, latitude, longitude, created_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, device_number, latitude, longitude, created_at
	`
	r.log.Debugf("Executing query: %s with values: %s, %f, %f, %v", query, position.DeviceNumber, position.Location.Latitude, position.Location.Longitude, position.CreatedAt)

	var newPosition models.Position
	err := r.conn.QueryRowContext(ctx, query, position.DeviceNumber, position.Location.Latitude, position.Location.Longitude, position.CreatedAt).Scan(
		&newPosition.ID,
		&newPosition.DeviceNumber,
		&newPosition.Location.Latitude,
		&newPosition.Location.Longitude,
		&newPosition.CreatedAt,
	)

	if err != nil {
		r.log.Errorf("Failed to create position: %v", err)
		return models.Position{}, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("Position created successfully: %+v", newPosition)
	return newPosition, nil
}

func (r *Repository) CreateOrUpdateCarsPosition(ctx context.Context, position models.CurrentPosition) (models.CurrentPosition, error) {
	query := `
		INSERT INTO cars_positions (id_company, id_car, latitude, longitude, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (id_car) 
		DO UPDATE SET 
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			updated_at = EXCLUDED.updated_at
		RETURNING id, id_company, id_car, latitude, longitude, updated_at
	`

	r.log.Debugf("Executing query: %s with values: %s, %s, %f, %f, %v",
		query, position.IDCompany, position.IDCar, position.Location.Latitude, position.Location.Longitude, position.UpdateAt)

	var newPosition models.CurrentPosition
	err := r.conn.QueryRowContext(ctx, query,
		position.IDCompany, position.IDCar, position.Location.Latitude, position.Location.Longitude, position.UpdateAt).Scan(
		&newPosition.ID,
		&newPosition.IDCompany,
		&newPosition.IDCar,
		&newPosition.Location.Latitude,
		&newPosition.Location.Longitude,
		&newPosition.UpdateAt,
	)

	if err != nil {
		r.log.Errorf("Failed to create or update car position: %v", err)
		return models.CurrentPosition{}, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("Car position created or updated successfully: %+v", newPosition)
	return newPosition, nil
}

// GetCarRoutePositions retrieves the positions of a car within a specific time range.
func (r *Repository) GetCarRoutePositions(ctx context.Context, carID string, from time.Time, to time.Time) ([]models.Position, error) {
	var positions []models.Position

	r.log.Debugf("Querying route positions for carID: %s from %v to %v", carID, from, to)

	query := `
	WITH car_info AS (
		SELECT device_number 
		FROM cars 
		WHERE id = $1
		LIMIT 1
	)
	SELECT id, device_number, latitude, longitude, created_at
	FROM position_data
		WHERE device_number = (select device_number from car_info)
		AND created_at BETWEEN $2 AND $3
		ORDER BY created_at ASC;
	`

	rows, err := r.conn.Query(query, carID, from, to)
	if err != nil {
		r.log.Errorf("Failed to execute query: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}
	defer rows.Close()

	for rows.Next() {
		var position models.Position

		if err := rows.Scan(&position.ID, &position.DeviceNumber, &position.Location.Latitude, &position.Location.Longitude, &position.CreatedAt); err != nil {
			r.log.Errorf("Failed to scan row: %v", err)
			return nil, fmt.Errorf("%w: %v", models.ErrFailedToProcessRow, err)
		}

		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorf("Error while iterating rows: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrRowsIterationError, err)
	}

	r.log.Debugf("Found %d positions for carID %s", len(positions), carID)

	return positions, nil
}

func (r *Repository) GetCurrentCarPositions(ctx context.Context, id string) ([]models.CurrentPositionResponse, error) {
	query := `
	SELECT 
		c.id,
		dt.latitude,
		dt.longitude,
		c.id_unicum
	FROM cars_positions AS dt
	JOIN cars AS c ON dt.id_car = c.id
	WHERE c.id_company = $1;
	`

	r.log.Debugf("Executing query to fetch current car positions for company_id=%s", id)

	rows, err := r.conn.Query(query, id)
	if err != nil {
		r.log.Errorf("%v: %v", models.ErrFailedToExecuteQuery, err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}
	defer rows.Close()

	var positions []models.CurrentPositionResponse

	for rows.Next() {
		var position models.CurrentPositionResponse
		if err := rows.Scan(
			&position.IDCar,
			&position.Point.Latitude,
			&position.Point.Longitude,
			&position.IDUnicum,
		); err != nil {
			r.log.Errorf("%v: %v", models.ErrFailedToProcessRow, err)
			return nil, fmt.Errorf("%w: %v", models.ErrFailedToProcessRow, err)
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorf("%v: %v", models.ErrRowsIterationError, err)
		return nil, fmt.Errorf("%w: %v", models.ErrRowsIterationError, err)
	}

	if len(positions) == 0 {
		r.log.Debugf("%v: No car positions found for company_id=%s", models.ErrNoContent, id)
		return nil, models.ErrNoContent
	}

	r.log.Debugf("Successfully fetched %d current car positions for company_id=%s", len(positions), id)

	return positions, nil
}

// GetCurrentCarPositionsByPoints retrieves cars in the area between two points.
func (r *Repository) GetCurrentCarPositionsByPoints(ctx context.Context, pointA models.Point, pointB models.Point) ([]models.CurrentPositionResponse, error) {
	var positions []models.CurrentPositionResponse

	r.log.Debugf("Querying car positions in area: [%f, %f] (lat) x [%f, %f] (lng)", pointA.Latitude, pointB.Latitude, pointA.Longitude, pointB.Longitude)

	if pointA.Latitude > pointB.Latitude {
		pointA.Latitude, pointB.Latitude = pointB.Latitude, pointA.Latitude
	}
	if pointA.Longitude > pointB.Longitude {
		pointA.Longitude, pointB.Longitude = pointB.Longitude, pointA.Longitude
	}

	query := `
		SELECT latitude, longitude, device_number
		FROM position_data
		WHERE latitude BETWEEN $1 AND $2
		AND longitude BETWEEN $3 AND $4
	`

	rows, err := r.conn.Query(query, pointA.Latitude, pointB.Latitude, pointA.Longitude, pointB.Longitude)
	if err != nil {
		r.log.Errorf("Failed to execute query: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}
	defer rows.Close()

	for rows.Next() {
		var position models.CurrentPositionResponse
		if err := rows.Scan(
			&position.Point.Latitude,
			&position.Point.Longitude,
			&position.IDCar,
		); err != nil {
			r.log.Errorf("Failed to scan row: %v", err)
			return nil, fmt.Errorf("%w: %v", models.ErrFailedToProcessRow, err)
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorf("Error while iterating rows: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrRowsIterationError, err)
	}

	r.log.Debugf("Found %d positions in the specified area.", len(positions))

	return positions, nil
}

// Breakage
// CreateBreakage inserts a new breakage record into the database and returns the created breakage ID.
func (r *Repository) CreateBreakage(ctx context.Context, breakage models.Breakage) (models.Breakage, error) {
	query := `
		INSERT INTO breakages (id_car, id_driver, latitude, longitude, type, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, id_car, id_driver, latitude, longitude, type, description, created_at`

	r.log.Debugf("Executing query to create breakage with values: car_id=%s, driver=%s, latitude=%f, longitude=%f, type=%s, description=%s, created_at=%v",
		breakage.CarID, breakage.DriverID, breakage.Location.Latitude, breakage.Location.Longitude, breakage.Type, breakage.Description, breakage.CreatedAt)

	var newBreakage models.Breakage
	err := r.conn.QueryRowContext(ctx, query,
		breakage.CarID,
		breakage.DriverID,
		breakage.Location.Latitude,
		breakage.Location.Longitude,
		breakage.Type,
		breakage.Description,
		breakage.CreatedAt).
		Scan(
			&newBreakage.ID,
			&newBreakage.CarID,
			&newBreakage.DriverID,
			&newBreakage.Location.Latitude,
			&newBreakage.Location.Longitude,
			&newBreakage.Type,
			&newBreakage.Description,
			&newBreakage.CreatedAt,
		)

	if err != nil {
		r.log.Errorf("Failed to create breakage: %v", err)
		return models.Breakage{}, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("Breakage created successfully with ID: %s", breakage.ID)
	return newBreakage, nil
}

// CreateBreakageFromMqtt processes the breakage data received from MQTT, inserts it into the database, and returns the created breakage record.
// func (r *Repository) CreateBreakageFromMqtt(ctx context.Context, breakage models.BreakageFromMqtt) (models.Breakage, error) {
// 	// datetime, err := time.Parse(time.RFC3339, breakage.CreatedAt)
// 	// if err != nil {
// 	// 	r.log.Errorf("Failed to parse created_at: %v", err)
// 	// 	return models.Breakage{}, fmt.Errorf("%w: %v", models.ErrFailedToProcessRow, err)
// 	// }

// 	// query := `
// 	// 	INSERT INTO breakages (id_car, id_driver, latitude, longitude, type, description, created_at)
// 	// 	VALUES ((SELECT id FROM cars WHERE device_number = $2 LIMIT 1), (SELECT id FROM cars where (SELECT id FROM cars WHERE device_number = $2 LIMIT 1),  $3, $4, $5, $6)
// 	// 	RETURNING id, id_car, latitude, longitude, type, description, created_at
// 	// `

// 	query := `
// 	WITH car_info AS (
//     SELECT id
//     FROM cars
//     WHERE device_number = $1
//     LIMIT 1
// 	),
// 	driver_info AS (
// 		SELECT id
// 		FROM drivers
// 		WHERE id_car = (SELECT id FROM car_info)
// 		LIMIT 1
// 	)
// 	INSERT INTO breakages (id_car, id_driver, latitude, longitude, type, description, created_at)
// 	VALUES (
// 		(SELECT id FROM car_info),
// 		(SELECT id FROM driver_info),
// 		$2, $3, $4, $5
// 	)
// 	RETURNING id, id_car, latitude, longitude, type, description, created_at;
// 	`

// 	r.log.Debugf("Executing query to create breakage with values: device_number=%s, latitude=%f, longitude=%f, type=%s, description=%s, created_at=%v",
// 		breakage.DeviceNum, breakage.Point[0], breakage.Point[1], breakage.Type, breakage.Description, breakage.CreatedAt)

// 	var createdBreakage models.Breakage

// 	err := r.conn.QueryRowContext(ctx, query,
// 		breakage.DeviceNum,
// 		breakage.Point[0],
// 		breakage.Point[1],
// 		breakage.Type,
// 		breakage.Description,
// 		breakage.CreatedAt,
// 	).Scan(
// 		&createdBreakage.ID,
// 		&createdBreakage.CarID,
// 		&createdBreakage.DriverID,
// 		&createdBreakage.Location.X,
// 		&createdBreakage.Location.Y,
// 		&createdBreakage.Type,
// 		&createdBreakage.Description,
// 		&createdBreakage.CreatedAt,
// 	)

// 	if err != nil {
// 		r.log.Errorf("Failed to create breakage: %v", err)
// 		return models.Breakage{}, fmt.Errorf("%w: %w", models.ErrFailedToExecuteQuery, err)
// 	}

// 	r.log.Debugf("Breakage created successfully: %+v", createdBreakage)
// 	return createdBreakage, nil
// }

func (r *Repository) CreateBreakageFromMqtt(ctx context.Context, breakage models.BreakageFromMqtt) (models.Breakage, error) {
	if breakage.DeviceNum == "" {
		r.log.Errorf("Device number is empty, cannot proceed with the operation")
		return models.Breakage{}, fmt.Errorf("device number cannot be empty")
	}

	// query := `
	// WITH car_info AS (
	// 	SELECT id
	// 	FROM cars
	// 	WHERE device_number = $1
	// 	LIMIT 1
	// ),
	// driver_info AS (
	// 	SELECT id
	// 	FROM drivers
	// 	WHERE id_car = (SELECT id FROM car_info)
	// 	LIMIT 1
	// )
	// INSERT INTO breakages (id_car, id_driver, latitude, longitude, type, description, created_at)
	// VALUES (
	// 	(SELECT id FROM car_info.id),
	// 	(SELECT id FROM driver_info.id),
	// 	$2, $3, $4, $5, $6
	// )
	// RETURNING id, id_car, id_driver, latitude, longitude, type, description, created_at;
	// `

	query := `
	WITH car_info AS (
			SELECT id 
			FROM cars 
			WHERE device_number = $1
			LIMIT 1
		),
		driver_info AS (
			SELECT id 
			FROM drivers
			WHERE id_car = (SELECT id FROM car_info)
			LIMIT 1
		)
	INSERT INTO breakages (id_car, id_driver, latitude, longitude, type, description, created_at)
	VALUES (
		(SELECT car_info.id FROM car_info),
		(SELECT driver_info.id FROM driver_info LIMIT 1), -- Если водителя нет, вставится NULL
		$2, $3, $4, $5, $6
	)
	RETURNING id, id_car, id_driver, latitude, longitude, type, description, created_at;
	`

	r.log.Debugf("Executing query to create breakage: device_number=%s, latitude=%f, longitude=%f, type=%s, description=%s, created_at=%v",
		breakage.DeviceNum, breakage.Point[0], breakage.Point[1], breakage.Type, breakage.Description, breakage.CreatedAt)

	var createdBreakage models.Breakage

	err := r.conn.QueryRowContext(ctx, query,
		breakage.DeviceNum,
		breakage.Point[0],
		breakage.Point[1],
		breakage.Type,
		breakage.Description,
		breakage.CreatedAt,
	).Scan(
		&createdBreakage.ID,
		&createdBreakage.CarID,
		&createdBreakage.DriverID,
		&createdBreakage.Location.Latitude,
		&createdBreakage.Location.Longitude,
		&createdBreakage.Type,
		&createdBreakage.Description,
		&createdBreakage.CreatedAt,
	)

	if err == sql.ErrNoRows {
		r.log.Debugf("%v: no breakage found for device_number=%s", models.ErrNoContent, breakage.DeviceNum)
		return models.Breakage{}, models.ErrNoContent
	}

	if err != nil {
		r.log.Errorf("%v: %v", models.ErrFailedToExecuteQuery, err)
		return models.Breakage{}, fmt.Errorf("%w: %w", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("Breakage created successfully: %+v", createdBreakage)
	return createdBreakage, nil
}

func (r *Repository) CheckDriverExists(ctx context.Context, deviceNumber string) (bool, error) {
	query := `
		WITH car_info AS (
			SELECT id
			FROM cars 
			WHERE device_number = $1
			LIMIT 1
		)
		SELECT EXISTS (
			SELECT 1
			FROM drivers
			WHERE id_car = (SELECT id FROM car_info)
		);
	`

	r.log.Debugf("Executing query to check driver existence for device_number: %s", deviceNumber)

	var driverExists bool
	err := r.conn.QueryRowContext(ctx, query, deviceNumber).Scan(&driverExists)
	if err != nil {
		r.log.Errorf("Failed to execute query: %v", err)
		return false, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	if driverExists {
		r.log.Debugf("Driver exists for device_number: %s", deviceNumber)
	} else {
		r.log.Debugf("No driver found for device_number: %s", deviceNumber)
	}

	return driverExists, nil
}

// Notification
func (r *Repository) CreateNotification(new models.Notification) (models.Notification, error) {
	query := `
		WITH car_info AS (
			SELECT id, id_company
			FROM cars 
			WHERE id = $1
			LIMIT 1
		)
        INSERT INTO notifications (id_user, id_breakages, note, status, created_at)
        VALUES ((SELECT id_company FROM car_info), $2, $3, $4, $5)
        RETURNING id, id_user, id_breakages, note, status, created_at`

	var createdNotification models.Notification
	err := r.conn.QueryRow(query,
		new.IDCar,
		new.IDBreakage,
		new.Note,
		new.Status,
		new.CreatedAt).
		Scan(
			&createdNotification.ID,
			&createdNotification.IDCar,
			&createdNotification.IDBreakage,
			&createdNotification.Note,
			&createdNotification.Status,
			&createdNotification.CreatedAt)

	if err != nil {
		return models.Notification{}, fmt.Errorf("error creating notification: %w", err)
	}

	return createdNotification, nil
}

func (r *Repository) UpdateNotificationStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE notifications
		SET status = $1
		WHERE id = $2`

	result, err := r.conn.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrNoContent
	}

	return nil
}

func (r *Repository) UpdateAllNotificationsStatus(ctx context.Context, userID string, status string) error {
	query := `
		UPDATE notifications
		SET status = $1
		WHERE id_user = $2`

	result, err := r.conn.ExecContext(ctx, query, status, userID)
	if err != nil {
		return fmt.Errorf("failed to execute update query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to fetch affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no notifications found for user")
	}

	return nil
}

// GetNotificationInfo retrieves detailed notification information from the database based on the notification ID.
func (r *Repository) GetNotificationInfo(ctx context.Context, notificationID string) (models.NotificationInfo, error) {
	query := `
	SELECT 
		n.note,
    	CONCAT(d.surname, d.name, d.middle_name) AS driver_name,
		b.latitude, 
		b.longitude,
		n.created_at
	FROM notifications n
	INNER JOIN breakages b ON n.id_breakages = b.id
	INNER JOIN drivers d ON b.id_driver = d.id
	WHERE n.id = $1;
	`

	r.log.Debugf("Executing query to fetch notification info for notificationID: %s", notificationID)

	var notificationInfo models.NotificationInfo
	err := r.conn.QueryRowContext(ctx, query, notificationID).Scan(
		&notificationInfo.Description,
		&notificationInfo.DriverName,
		&notificationInfo.Location.Latitude,
		&notificationInfo.Location.Longitude,
		&notificationInfo.CreatedAt,
	)

	if err == sql.ErrNoRows {
		r.log.Errorf("No rows found for notificationID: %s", notificationID)
		return models.NotificationInfo{}, models.ErrNoContent
	}

	if err != nil {
		r.log.Errorf("Failed to execute query for notificationID: %s, error: %v", notificationID, err)
		return models.NotificationInfo{}, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}

	r.log.Debugf("Successfully fetched notification info for notificationID: %s", notificationID)
	return notificationInfo, nil
}

// GetNotificationList retrieves a list of notifications based on the provided status, limit, and offset.
func (r *Repository) GetNotificationList(ctx context.Context, userID string, status *string, limit, offset int) ([]models.NotificationListItem, error) {
	query := `
		SELECT 
			n.id,
			c.state_number,
			c.brand,
			b.type AS breakage_type,
			n.created_at
		FROM notifications n
		INNER JOIN breakages b ON n.id_breakages = b.id
		INNER JOIN cars c ON b.id_car = c.id
		WHERE n.id_user = $1 AND n.status = COALESCE($2, n.status)
		ORDER BY n.created_at DESC
		LIMIT $3 OFFSET $4`

	r.log.Debugf("Executing query to fetch notifications with user_id: %s status: %v, limit: %d, offset: %d", userID, status, limit, offset)

	rows, err := r.conn.QueryContext(ctx, query, userID, status, limit, offset)
	if err != nil {
		r.log.Errorf("Failed to execute query: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}
	defer rows.Close()

	var notifications []models.NotificationListItem
	for rows.Next() {
		var item models.NotificationListItem
		if err := rows.Scan(
			&item.ID,
			&item.StateNumber,
			&item.Brand,
			&item.BreakageType,
			&item.CreatedAt,
		); err != nil {
			r.log.Errorf("Failed to scan row: %v", err)
			return nil, fmt.Errorf("%w: %v", models.ErrFailedToProcessRow, err)
		}
		notifications = append(notifications, item)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorf("Error while iterating rows: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrRowsIterationError, err)
	}

	r.log.Debugf("Successfully fetched %d notifications", len(notifications))
	return notifications, nil
}

// Report
func (r *Repository) GetReportData(userId string) ([]models.ReportData, error) {
	query := `
		SELECT
			w.id AS wheel_id,             
			c.state_number,              
			w.brand AS tire_brand,       
			w.mileage,                   
			COUNT(CASE WHEN s.temperature < w.min_temperature 
				OR s.temperature > w.max_temperature THEN 1 END) 
			AS temp_out_of_bounds,  
			COUNT(CASE WHEN s.pressure < w.min_pressure 
				OR s.pressure > w.max_pressure THEN 1 END) 
			AS pressure_out_of_bounds  
		FROM
			cars c
		JOIN wheels w ON w.id_car = c.id
		JOIN sensors_data s ON s.device_number = c.device_number 
			AND s.sensor_number = w.sensor_number
		WHERE
			c.id_company = $1  
		GROUP BY
			w.id, c.state_number, w.brand, w.mileage, w.position  
		ORDER BY
			c.state_number, w.position; 
	`

	r.log.Debugf("Executing query: %s with userId: %s", query, userId)

	rows, err := r.conn.Query(query, userId)
	if err != nil {
		r.log.Errorf("Failed to execute query: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrFailedToExecuteQuery, err)
	}
	defer rows.Close()

	var reportData []models.ReportData

	for rows.Next() {
		var data models.ReportData
		if err := rows.Scan(
			&data.IdWheel,
			&data.StateNumber,
			&data.TireBrand,
			&data.Mileage,
			&data.TempOutOfBounds,
			&data.PressureOutOfBounds,
		); err != nil {
			r.log.Errorf("Failed to scan row: %v", err)
			return nil, fmt.Errorf("%w: %v", models.ErrFailedToProcessRow, err)
		}
		reportData = append(reportData, data)
	}

	if err := rows.Err(); err != nil {
		r.log.Errorf("Rows iteration error: %v", err)
		return nil, fmt.Errorf("%w: %v", models.ErrRowsIterationError, err)
	}

	r.log.Debugf("Successfully retrieved %d report records", len(reportData))
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
