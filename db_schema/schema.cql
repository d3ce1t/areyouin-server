// Drop and recreate the keyspace
DROP KEYSPACE IF EXISTS areyouin;
CREATE KEYSPACE areyouin WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

// Create the schema
USE areyouin;

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
	PRIMARY KEY (user_id, end_date, event_id) // FIXME: Make partitions because a user could have a lot of events as time goes further
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (end_date ASC, event_id ASC);

// TODO: Study event partitions by date
DROP TABLE IF EXISTS event;
CREATE TABLE event (
	// Partition Key
	event_id bigint,
	// Event Info
	author_id bigint STATIC,
	author_name text STATIC,
	message text STATIC,
  created_date timestamp STATIC,
	start_date timestamp STATIC,
	end_date timestamp STATIC,
	public boolean STATIC,
	latitude float STATIC,
	longitude float STATIC,
	num_attendees int STATIC, // maybe use a counter?
	num_guests int STATIC,
	// Participants
	guest_id bigint,
	guest_name text, // participant name
	guest_response int, // 0) no response, 1) no assist, 2) cannot assist, 3) assist
	guest_status int, // 0) no delivered, 1) server, 2) client
	PRIMARY KEY (event_id, guest_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};
//AND CLUSTERING ORDER BY (event_id DESC);
