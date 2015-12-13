// Drop and recreate the keyspace
DROP KEYSPACE IF EXISTS areyouin;
CREATE KEYSPACE areyouin WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };

// Create the schema
USE areyouin;

CREATE TABLE user_email_credentials (
	email text,
	password text,
	user_id bigint,
	PRIMARY KEY (email)
);

CREATE TABLE user_facebook_credentials (
	fb_id text,
	fb_token text,
	user_id bigint,
	PRIMARY KEY (fb_id)
);

CREATE TABLE user_account (
	user_id bigint,
	auth_token UUID,
	email text,
	email_verified boolean,
	name text,
	phone text,
	phone_verified boolean,
	fb_id text text,
	fb_token text text,
	last_connection timestamp,
	created_date timestamp,
	PRIMARY KEY (user_id)
);

CREATE TABLE user_friend (
	user_id bigint,
	group_id int,
	group_name text STATIC,
	friend_id bigint,
	name text,
	PRIMARY KEY ((user_id, group_id), friend_id)
);

// look up table
CREATE TABLE user_events (
	user_id bigint,
	event_id bigint,
	// author_id bigint,
	// author_name text,
	// image,
	start_date timestamp,
	end_date timestamp,
	response int, // 0) no response, 1) no assist, 2) cannot assist, 3) assist
	// num_attendees int,
	PRIMARY KEY (user_id, event_id)
)
WITH CLUSTERING ORDER BY (start_date DESC, end_date ASC);

// TODO: Study event partitions by date
CREATE TABLE event (
	event_id bigint,
	author_id bigint,
	author_name text STATIC,
	message text,
	start_date timestamp,
	end_date timestamp,
	public boolean,
	latitude float,
	longitude float,
	num_attendees int, // maybe use a counter?
	num_guests int,
	created_date timestamp,
	PRIMARY KEY (author_id, event_id)
)
WITH CLUSTERING ORDER BY (start_date DESC, end_date ASC);

CREATE TABLE participant (
	event_id bigint,
	user_id bigint,
	name text, // participant name
	response int, // 0) no response, 1) no assist, 2) cannot assist, 3) assist
	status int, // 0) no delivered, 1) server, 2) client
	PRIMARY KEY (event_id, user_id)
);
