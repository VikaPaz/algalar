CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	inn int,
	name varchar(100),
	surname varchar(100),
	middle_name varchar(100),
	login varchar(100),
	password varchar(100),
	timezone text
);

CREATE TABLE IF NOT EXISTS cars (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	id_company uuid references users,
	state_namber varchar(100),
	brand varchar(100),
	id_device int,
	id_unicum int,
	count_axis int
);

CREATE TABLE IF NOT EXISTS wheels (
	id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
	id_car uuid references cars,
	axis_number int,
	position int,
	size int,
	cost int,
	brand varchar(100),
	model varchar(100),
	mileage int,
	min_temperature int,
	min_pressure int,
	max_temperature int,
	max_pressure int
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
	id_company uuid REFERENCES users,
	id_car uuid REFERENCES cars,
	id_driver uuid REFERENCES drivers,
	type varchar(100),
	discription varchar(100),
	datetime timestamp
);