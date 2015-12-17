// TODO: Study size of commitlogs in order to avoid growing so much

// Drop and recreate the keyspace
DROP KEYSPACE IF EXISTS areyouin_demo;
CREATE KEYSPACE areyouin_demo WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 3 };

// Create the schema
USE areyouin_demo;

DROP TABLE IF EXISTS user_email_credentials;
CREATE TABLE user_email_credentials (
	email text,
	user_id bigint,
	password blob, // 32 Bytes
	salt blob, // 32 bytes
	PRIMARY KEY (email)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

DROP TABLE IF EXISTS user_facebook_credentials;
CREATE TABLE user_facebook_credentials (
	fb_id text,
	fb_token text,
	user_id bigint,
	PRIMARY KEY (fb_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

DROP TABLE IF EXISTS user_account;
CREATE TABLE user_account (
	user_id bigint,
	auth_token UUID,
	email text,
	email_verified boolean,
	name text,
	phone text,
	phone_verified boolean,
	fb_id text,
	fb_token text,
	last_connection timestamp,
	created_date timestamp,
	PRIMARY KEY (user_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

DROP TABLE IF EXISTS user_friends;
CREATE TABLE user_friends (
	user_id bigint,
	group_id int,
	group_name text STATIC,
	friend_id bigint,
	name text,
	PRIMARY KEY ((user_id, group_id), friend_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// look up table
DROP TABLE IF EXISTS user_events;
CREATE TABLE user_events (
	user_id bigint,
	// author_id bigint,
	// author_name text,
	// image,
	end_date timestamp,
	// start_date timestamp,
	event_id bigint,
	response int, // 0) no response, 1) no assist, 2) cannot assist, 3) assist
	// num_attendees int,
	PRIMARY KEY (user_id, end_date, event_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (end_date ASC, event_id ASC);

// TODO: Study event partitions by date
DROP TABLE IF EXISTS event;
CREATE TABLE event (
	event_id bigint,
	author_id bigint,
	author_name text, // STATIC,
	message text,
	start_date timestamp,
	end_date timestamp,
	public boolean,
	latitude float,
	longitude float,
	num_attendees int, // maybe use a counter?
	num_guests int,
	created_date timestamp,
	PRIMARY KEY (event_id) // removed author_id as partition key
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};
//AND CLUSTERING ORDER BY (event_id DESC);

DROP TABLE IF EXISTS event_participants;
CREATE TABLE event_participants (
	event_id bigint,
	user_id bigint,
	name text, // participant name
	response int, // 0) no response, 1) no assist, 2) cannot assist, 3) assist
	status int, // 0) no delivered, 1) server, 2) client
	PRIMARY KEY (event_id, user_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};
