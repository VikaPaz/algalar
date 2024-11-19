CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	inn varchar(100),
	name varchar(100),
	surname varchar(100),
	middle_name varchar(100),
	login varchar(100),
	password varchar(100),
	post varchar(100),
	timezone text
);

CREATE TABLE IF NOT EXISTS cars (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	id_company uuid REFERENCES users,
	state_number varchar(100),
	brand varchar(100),
	id_device varchar(100),
	id_unicum varchar(100),
	count_axis int
);

CREATE TABLE IF NOT EXISTS wheels (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	id_company uuid REFERENCES users,
	id_car uuid REFERENCES cars,
	count_axis int,
	position int,
	size float,
	cost float,
	brand varchar(100),
	model varchar(100),
	ngp float,
	tkvh float,
	mileage float,
	min_temperature float,
	min_pressure float,
	max_temperature float,
	max_pressure float
);

CREATE TABLE IF NOT EXISTS drivers (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	id_company uuid REFERENCES users,
	id_car uuid REFERENCES cars,
	name varchar(100),
	surname varchar(100),
	middle_name varchar(100),
	experience varchar(100),
	phone varchar(100),
	birthday timestamp,
	road varchar(100),
	score varchar(100)
);


CREATE TABLE IF NOT EXISTS breakages (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	car_id uuid REFERENCES cars,
	state_number varchar(100),
	type varchar(100),
	description varchar(100),
	datetime timestamp
);
-- TODO: add time
CREATE TABLE IF NOT EXISTS sensors (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	car_id uuid REFERENCES cars,
	state_number varchar(100),
	count_axis int,
	position int,
	pressure float,
	temperature float,
	datetime timestamp
);

CREATE TABLE IF NOT EXISTS notifications (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	user_id uuid REFERENCES users,
	breakages_id uuid REFERENCES breakages,
	status varchar(100)
);