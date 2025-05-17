package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
)

// SQLite config for reads and writes (avoid SQLITE BUSY error): https://kerkour.com/sqlite-for-servers
func openAppDB(dbpath string) (*sql.DB, bool, error) {
	// Create string of connection params to prevent "SQLITE_BUSY" errors - to be further improved based on the above article
	dbParams := "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"

	// Either connect to or create (if it doesn't exist) the database at the provided path
	db, err := sql.Open("sqlite3", dbpath+dbParams)
	if err != nil {
		return nil, false, err
	}

	// Create context with 5 second deadline so that we can ping the db and finish establishing a db connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, false, err
	}

	// Set up logs table
	err = createLogsTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up users table
	err = createUserTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up tokens table
	err = createTokenTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up config table
	err = createConfigTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up listener_events table
	err = createListenerEventsTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Check to see whether a user already exists in the database - if not, a user must be created before the admin dashboard can be used
	exists, err := userExists(db)
	if err != nil {
		log.Fatal(err)
	}

	return db, exists, nil
}

// SQLite config for reads and writes (avoid SQLITE BUSY error): https://kerkour.com/sqlite-for-servers
func openKamarDB(dbpath string) (*sql.DB, error) {
	// Create string of connection params to prevent "SQLITE_BUSY" errors - to be further improved based on the above article
	dbParams := "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL"

	// Either connect to or create (if it doesn't exist) the database at the provided path
	db, err := sql.Open("sqlite3", dbpath+dbParams)
	if err != nil {
		return nil, err
	}

	// Create context with 5 second deadline so that we can ping the db and finish establishing a db connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Set up tables for data to be consumed from SMS
	err = createSMSTables(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createLogsTable(db *sql.DB) error {
	userTableStmt := `CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level TEXT NOT NULL,
		time TEXT NOT NULL DEFAULT (datetime('now')),
		message TEXT NOT NULL,
		properties TEXT,
		trace TEXT
	);`

	_, err := db.Exec(userTableStmt)

	if err != nil {
		return err
	}

	// Create an indexed column based on any logs that come with an attached user id, to make it easier to query for logs regarding a specific user. VIRTUAL allows the column to store NULL values without errors, and NULL values are ignored in indexes
	// Check the datatype of $.user_id before extracting, as otherwise in some cases the value of this column can be set to a single " and cause errors
	// NOTE: Even with the below query, properties with a user_id value of "" still cause malformed JSON errors (specifically the logs created by logsPage query params)
	alterTableStmt := `ALTER TABLE logs ADD COLUMN user_id INTEGER
	GENERATED ALWAYS AS (
    CASE 
			WHEN json_valid(properties) 
				AND json_type(json_extract(properties, '$.user_id')) = 'integer' 
				AND json_extract(properties, '$.user_id') != '' 
			THEN json_extract(properties, '$.user_id') 
			ELSE NULL 
    END
	) VIRTUAL;`

	_, err = db.Exec(alterTableStmt)
	// Alter table doesn't support IF NOT EXISTS, so ignore the error thrown if this column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	createIndexStmt := `
		CREATE INDEX IF NOT EXISTS idx_logs_user_id ON logs(user_id);
	`

	_, err = db.Exec(createIndexStmt)

	// FTS5 tables are optimised for text search, and allow in this case for more efficiently searching for logs containing key words
	// This particular FTS5 table is only set up to allow for searching for text in log messages only, as defined by the message param
	createFTSTableStmt := `
		CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts
		USING fts5(message, content='logs', content_rowid='id');
	`

	_, err = db.Exec(createFTSTableStmt)

	createLogsMetadataTableStmt := `
		CREATE TABLE IF NOT EXISTS logs_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			level TEXT UNIQUE,
			user_id INTEGER UNIQUE,
			count INTEGER NOT NULL DEFAULT 0
		);`

	_, err = db.Exec(createLogsMetadataTableStmt)

	return err
}

func createUserTable(db *sql.DB) error {
	userTableStmt := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		last_authenticated_at TEXT NOT NULL DEFAULT (datetime('now')),
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);`

	_, err := db.Exec(userTableStmt)

	return err
}

func createTokenTable(db *sql.DB) error {
	tokenTableStmt := `CREATE TABLE IF NOT EXISTS tokens (
		hash BLOB PRIMARY KEY, 
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
		expiry TEXT NOT NULL
	);`

	_, err := db.Exec(tokenTableStmt)

	return err
}

// id INTEGER PRIMARY KEY AUTOINCREMENT,
// key TEXT UNIQUE NOT NULL,
func createConfigTable(db *sql.DB) error {
	configTableStmt := `CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		type TEXT NOT NULL,
		description TEXT NOT NULL,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(configTableStmt)

	// Set default values for config - ignore if they already exist
	configTableStmt = `
	INSERT OR IGNORE INTO config (key, value, type, description) VALUES 
		('kamar_ip', '192.168.1.1', 'string', "IP address of your school's instance of KAMAR - find by running ifconfig"),
		('service_name', 'KAMAR Listener Service', 'string', 'Use the acronym/name of your school, eg. "WHS KAMAR Listener Service"'),
		('info_url', 'https://www.educationcounts.govt.nz/directories/list-of-nz-schools', 'string', 'Website where people can contact you/read about how you use this service, eg. https://schoolname.school.nz'),
		('privacy_statement', 'This service only collects results data, and stores it locally on a secure device. Only staff members of the school have access to the data.', 'string', 'Minimum 100 characters: a description of how you use the data from this listener service'),
		('listener_username', '', 'string', 'Username entered into KAMAR when setting up listener service'),
		('listener_password', '', 'password', 'Password entered into KAMAR when setting up listener service'),
		('details', 'true', 'bool', 'Enable/disable details'),
		('passwords', 'true', 'bool', 'Enable/disable passwords'),
		('photos', 'true', 'bool', 'Enable/disable photos'),
		('groups', 'true', 'bool', 'Enable/disable groups'),
		('awards', 'true', 'bool', 'Enable/disable awards'),
		('timetables', 'true', 'bool', 'Enable/disable timetables'),
		('attendance', 'true', 'bool', 'Enable/disable attendance'),
		('assessments', 'true', 'bool', 'Enable/disable results and assessments'),
		('pastoral', 'true', 'bool', 'Enable/disable pastoral'),
		('learning_support', 'true', 'bool', 'Enable/disable learning support'),
		('subjects', 'true', 'bool', 'Enable/disable subjects'),
		('notices', 'true', 'bool', 'Enable/disable notices'),
		('calendar', 'true', 'bool', 'Enable/disable calendar'),
		('bookings', 'true', 'bool', 'Enable/disable bookings');
	`

	_, err = db.Exec(configTableStmt)

	// configMetaTableStmt := `CREATE TABLE IF NOT EXISTS config_metadata (
	// 	config_id INTEGER PRIMARY KEY,
	// 	description TEXT NOT NULL,
	// 	updated_at TEXT DEFAULT (datetime('now')),
	// 	FOREIGN KEY (config_id) REFERENCES config(id) ON DELETE CASCADE
	// );`

	// _, err = db.Exec(configMetaTableStmt)

	return err
}

func createListenerEventsTable(db *sql.DB) error {
	listenerEventsTableStmt := `CREATE TABLE IF NOT EXISTS listener_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		req_type TEXT NOT NULL,
		was_successful INTEGER NOT NULL,
		message TEXT,
		time TEXT NOT NULL DEFAULT (datetime('now'))
	);`

	_, err := db.Exec(listenerEventsTableStmt)

	return err
}

func createSMSTables(db *sql.DB) error {
	// Includes resultData and results fields
	resultTableStmt := `CREATE TABLE IF NOT EXISTS results (
		code			TEXT,
		comment         TEXT,
		course          TEXT,
		curriculumlevel,
		date            TEXT,
		enrolled		INTEGER,
		id              INTEGER,
		nsn             TEXT,
		number          TEXT,
		published		INTEGER,
		result          TEXT,
		resultData TEXT,
		results TEXT,
		subject         TEXT NULL,
		tnv 			TEXT,
		type            TEXT,
		version         INTEGER,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		year            INTEGER,
		yearlevel       INTEGER,
		UNIQUE(id, tnv, subject)
	);`

	_, err := db.Exec(resultTableStmt)
	if err != nil {
		log.Printf("couldn't set up results table: %v", err)
	}

	assessmentTableStmt := `CREATE TABLE IF NOT EXISTS assessments (
		credits			INTEGER,
		description TEXT,
		internalexternal TEXT,
		level INTEGER,
		number TEXT,
		points TEXT,
		purpose TEXT,
		schoolref TEXT,
		subfield TEXT,
		title TEXT,
		tnv TEXT PRIMARY KEY,
		type TEXT,
		version INTEGER,
		weighting TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);`

	_, err = db.Exec(assessmentTableStmt)

	attendanceTableStmt := `CREATE TABLE IF NOT EXISTS attendance (
		student_id INTEGER PRIMARY KEY,
		nsn TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);`

	_, err = db.Exec(attendanceTableStmt)

	// Secondary table for attendance values joined on attendance table's row ID - necessary as values is an array
	attendanceValuesTableStmt := `CREATE TABLE IF NOT EXISTS attendance_values (
		att_student_id INTEGER NOT NULL,
		date TEXT,
		codes TEXT,
		alt TEXT,
		hdu INTEGER,
		hdj INTEGER,
		hdp INTEGER,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (att_student_id) REFERENCES attendance(student_id) ON DELETE CASCADE,
		UNIQUE(att_student_id, date)
	);`

	_, err = db.Exec(attendanceValuesTableStmt)

	// While there are performance benefits to creating a secondary effortsIndices table as below, updates will be safer and more consistent if they are held in an array in the base classEfforts table
	// TODO: efforts - how does SQLite best store arrays (of integers?)
	classEffortsTableStmt := `CREATE TABLE IF NOT EXISTS class_efforts (
		count INTEGER,
		student_id INTEGER NOT NULL,
		nsn TEXT,
		date TEXT NOT NULL,
		slot INTEGER NOT NULL,
		term INTEGER,
		week INTEGER,
		subject TEXT,
		user TEXT,
		efforts TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(student_id, date, slot)
	);`

	_, err = db.Exec(classEffortsTableStmt)

	// effortIndicesTableStmt := `CREATE TABLE IF NOT EXISTS effort_indices (
	// 	eff_student_id INTEGER NOT NULL,
	// 	date TEXT,
	// 	slot INTEGER,
	// 	value INTEGER,
	// 	listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	// 	FOREIGN KEY (eff_student_id) REFERENCES class_efforts(student_id) ON DELETE CASCADE,
	// 	UNIQUE(att_student_id, date, slot)
	// );`

	// _, err = db.Exec(effortIndicesTableStmt)

	pastoralTableStmt := `CREATE TABLE IF NOT EXISTS pastoral (
		student_id			INTEGER NOT NULL,
		nsn TEXT,
		type TEXT,
		ref INTEGER,
		reason TEXT,
		reason_pb TEXT,
		motivation TEXT,
		motivation_pb TEXT,
		location TEXT,
		location_pb TEXT,
		others_involved TEXT,
		action1 TEXT,
		action2 TEXT,
		action3 TEXT,
		action_pb1 TEXT,
		action_pb2 TEXT,
		action_pb3 TEXT,
		teacher TEXT,
		points INTEGER,
		demerits INTEGER,
		dateevent TEXT,
		timeevent TEXT,
		datedue TEXT,
		duestatus TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(student_id, type, ref)
	);`

	_, err = db.Exec(pastoralTableStmt)

	// TODO: values - how does SQLite best store arrays (of integers?)
	recognitionsTableStmt := `CREATE TABLE IF NOT EXISTS class_efforts (
		count INTEGER,
		student_id INTEGER NOT NULL,
		nsn TEXT,
		uuid TEXT
		date TEXT NOT NULL,
		slot INTEGER NOT NULL,
		term INTEGER,
		week INTEGER,
		subject TEXT,
		user TEXT,
		points INTEGER,
		comment TEXT,
		values TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(student_id, date, slot)
	);`

	_, err = db.Exec(recognitionsTableStmt)

	// TODO: Consider creating a singular "groups" table with a polymorphic foreign key - as it is, groups don't have an ID anyway, so would it be useful?
	// TODO: Consider joining on id rather than uniqueid - though id is text rather than numerical (it's the teacher code, eg. MJE), each teacher is guaranteed to have one? Or perhaps students should be joined to other tables on uniqueid?
	staffTableStmt := `CREATE TABLE IF NOT EXISTS staff (
		id TEXT,
		uuid TEXT PRIMARY KEY,
		role TEXT,
		created INTEGER,
		uniqueid INTEGER,
		username TEXT,
		firstname TEXT,
		lastname TEXT,
		gender TEXT,
		schoolindex INTEGER,
		title TEXT,
		email TEXT,
		mobile TEXT,
		extension TEXT,
		classification TEXT,
		position TEXT,
		house TEXT,
		tutor TEXT,
		datebirth TEXT,
		leavingdate TEXT,
		startingdate TEXT,
		eslguid TEXT,
		moenumber TEXT,
		photocopierid TEXT,
		registrationnumber TEXT,
		custom TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	
	CREATE TABLE IF NOT EXISTS staff_groups (
		staff_uuid TEXT NOT NULL,
		staff_id TEXT NOT NULL,
		type TEXT,
		subject TEXT,
		coreoption TEXT,
		ref INTEGER,
		year INTEGER,
		name TEXT,
		description TEXT,
		teacher TEXT,
		showreport INTEGER,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (staff_uuid) REFERENCES staff(uuid) ON DELETE CASCADE
	);`

	_, err = db.Exec(staffTableStmt)

	studentTableStmt := `CREATE TABLE IF NOT EXISTS students (
		id INTEGER NOT NULL,
		uuid TEXT PRIMARY KEY,
		role TEXT,
		created INTEGER,
		uniqueid INTEGER,
		nsn TEXT,
		username TEXT,
		firstname TEXT,
		firstnamelegal TEXT,
		lastname TEXT,
		lastnamelegal TEXT,
		forenames TEXT,
		forenameslegal TEXT,
		gender TEXT,
		genderpreferred TEXT,
		gendercode INTEGER,
		schoolindex INTEGER,
		email TEXT,
		mobile TEXT,
		house TEXT,
		whanau TEXT,
		boarder TEXT,
		byodinfo TEXT,
		ece TEXT,
		esol TEXT,
		ors TEXT,
		languagespoken TEXT,
		datebirth INTEGER,
		startingdate INTEGER,
		startschooldate TEXT,
		leavingdate INTEGER,
		leavingreason TEXT,
		leavingschool TEXT,
		leavingactivity TEXT,
		moetype TEXT,
		ethnicityL1 TEXT,
		ethnicityL2 TEXT,
		ethnicity TEXT,
		iwi TEXT,
		yearlevel TEXT,
		fundinglevel TEXT,
		tutor TEXT,
		timetablebottom1 TEXT,
		timetablebottom2 TEXT,
		timetablebottom3 TEXT,
		timetablebottom4 TEXT,
		timetabletop1 TEXT,
		timetabletop2 TEXT,
		timetabletop3 TEXT,
		timetabletop4 TEXT,
		maorilevel TEXT,
		pacificlanguage TEXT,
		pacificlevel TEXT,
		siblinglink TEXT,
		photocopierid TEXT,
		signedagreement TEXT,
		accountdisabled TEXT,
		networkaccess TEXT,
		altdescription TEXT,
		althomedrive TEXT,
		custom  TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS student_awards (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		type TEXT,
		name TEXT,
		year INTEGER,
		date TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE,
		UNIQUE(student_uuid, name, year, date)
	);
	
	CREATE TABLE IF NOT EXISTS student_caregivers (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		ref INTEGER,
		role TEXT,
		name TEXT,
		email TEXT,
		mobile TEXT,
		relationship TEXT,
		status TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE,
		UNIQUE(student_uuid, ref)
	);
	
	CREATE TABLE IF NOT EXISTS student_datasharing (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		details INTEGER,
		photo INTEGER,
		other INTEGER,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE,
		UNIQUE(student_uuid)
	);
	
	CREATE TABLE IF NOT EXISTS student_emergency (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		name TEXT,
		relationship TEXT,
		mobile TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS student_flags (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		general TEXT,
		notes TEXT,
		alert TEXT,
		conditions TEXT,
		dietary TEXT,
		ibuprofen TEXT,
		medical TEXT,
		paracetamol TEXT,
		pastoral TEXT,
		reactions TEXT,
		specialneeds TEXT,
		vaccinations TEXT,
		eotcconsent TEXT,
		eotcform TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE,
		UNIQUE(student_uuid)
	);

	CREATE TABLE IF NOT EXISTS student_groups (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		type TEXT,
		subject TEXT,
		coreoption TEXT,
		ref INTEGER,
		year INTEGER,
		name TEXT,
		description TEXT,
		teacher TEXT,
		showreport INTEGER,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS student_residences (
		student_uuid TEXT NOT NULL,
		student_id INTEGER NOT NULL,
		title TEXT,
		salutation TEXT,
		email TEXT,
		numFlatUnit TEXT,
		numStreet TEXT,
		ruralDelivery TEXT,
		suburb TEXT,
		town TEXT,
		postcode TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		FOREIGN KEY (student_uuid) REFERENCES students(uuid) ON DELETE CASCADE
	);`

	_, err = db.Exec(studentTableStmt)

	subjectTableStmt := `CREATE TABLE IF NOT EXISTS subjects (
		id TEXT PRIMARY KEY,
		created INTEGER,
		name TEXT,
		department TEXT,
		subdepartment TEXT,
		qualification TEXT,
		level INTEGER,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);`

	_, err = db.Exec(subjectTableStmt)

	// TODO: Consider changing "uuid" to "student_uuid" for clarity
	// TODO: Split into staff timetables and student timetables, seeing as they come as two separate sync types?
	timetablesTableStmt := `CREATE TABLE IF NOT EXISTS timetables (
		student INTEGER,
		uuid TEXT PRIMARY KEY,
		grid TEXT,
		timetable TEXT,
		listener_updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);`

	_, err = db.Exec(timetablesTableStmt)

	return err
}

func userExists(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count > 0, err
}
