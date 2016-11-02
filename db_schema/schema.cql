// Drop and recreate the keyspace
DROP KEYSPACE IF EXISTS areyouin;
CREATE KEYSPACE areyouin WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

// Create the schema
USE areyouin;

// Q1: Find a user with a known e-mail address
// Used by: NewAuthToken
DROP TABLE IF EXISTS user_email_credentials;
CREATE TABLE user_email_credentials (
	email text,
	user_id bigint,
	password blob, // 32 Bytes
	salt blob, // 32 bytes
	PRIMARY KEY (email)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q2: Find a user with a known Facebook ID
// Used by: NewAuthToken
DROP TABLE IF EXISTS user_facebook_credentials;
CREATE TABLE user_facebook_credentials (
	fb_id text,
	fb_token text,
	user_id bigint,
	created_date timestamp,
	PRIMARY KEY (fb_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q3: Find a user with a known UID
// Used by: UserAuth
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
	iid_token text,
	network_version int,
	platform text,
	last_connection timestamp,
	created_date timestamp,
	profile_picture blob,
	picture_digest blob,
	PRIMARY KEY (user_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q4: Find all friends by a given user_id
DROP TABLE IF EXISTS friends_by_user;
CREATE TABLE friends_by_user (
   user_id bigint,
   friend_id bigint,
   friend_name text,
   picture_digest blob,
   PRIMARY KEY (user_id, friend_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q5: Get all groups of a given user_id
DROP TABLE IF EXISTS groups_by_user;
CREATE TABLE groups_by_user (
	user_id bigint,
	group_id int,
	group_name text,
	group_size int,
	PRIMARY KEY (user_id, group_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q6: Get all friends inside a group by a given user_id and group_id
DROP TABLE IF EXISTS friends_by_group;
CREATE TABLE friends_by_group (
   user_id bigint,
	 group_id int,
	 friend_id bigint,
   PRIMARY KEY (user_id, group_id, friend_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q7: Find events a user was invited and that is finished (order by start_date DESC)
DROP TABLE IF EXISTS events_history_by_user;
CREATE TABLE events_history_by_user (
	user_id bigint,
	position timestamp,
	event_id bigint,
	PRIMARY KEY (user_id, position, event_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (position DESC, event_id DESC);

// Q8: Find a event with a specified event id
DROP TABLE IF EXISTS event;
CREATE TABLE event (
	// Partition Key
	event_id bigint,
	// Event Info
	author_id bigint STATIC,
	author_name text STATIC,
	message text STATIC,
	picture blob STATIC,
	picture_digest blob STATIC,
	created_date timestamp STATIC,
	inbox_position timestamp STATIC,
	start_date timestamp STATIC,
	end_date timestamp STATIC,
	public boolean STATIC,
	latitude float STATIC,
	longitude float STATIC,
	event_state int STATIC, // 0) Not started, 1) Ongoing, 2) Finished, 3) Cancelled
	event_timestamp bigint STATIC,
	// Participants
	guest_id bigint,
	guest_name text, // participant name
	guest_response int, // 0) no response, 1) no assist, 2) cannot assist, 3) assist
	guest_status int, // 0) no delivered, 1) server, 2) client
	PRIMARY KEY (event_id, guest_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q9: Find events that are in progress or not started (ordered by position ASC)
DROP TABLE IF EXISTS events_timeline;
CREATE TABLE events_timeline (
	// Partition key
	bucket int,
	//  Event Info	
	event_id bigint,
	position timestamp,
	PRIMARY KEY (bucket, position, event_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (position ASC, event_id ASC);

// Q10: Find a user or event thumbnail by his or her id and dpi
DROP TABLE IF EXISTS thumbnails;
CREATE TABLE thumbnails (
	id bigint,
	digest blob STATIC,
	dpi int,
	thumbnail blob,
	created_date timestamp,
	PRIMARY KEY (id, dpi)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q11: Find user access token by user_id
DROP TABLE IF EXISTS user_access_token;
CREATE TABLE user_access_token (
	user_id bigint,
	access_token UUID,
	last_used timestamp,
	created_date timestamp,
	PRIMARY KEY (user_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

// Q12: Get all friend requests received by an user and ordered by date
DROP TABLE IF EXISTS friend_requests_received;
CREATE TABLE friend_requests_received (
	user_id bigint,
	created_date timestamp,
	friend_id bigint,
	name text,
	email text,
	PRIMARY KEY (user_id, created_date, friend_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (created_date DESC, friend_id ASC);

// Q13: Find a friend request sent by user_id and friend_id
DROP TABLE IF EXISTS friend_requests_sent;
CREATE TABLE friend_requests_sent (
	user_id bigint,
	friend_id bigint,
	created_date timestamp,
	PRIMARY KEY (user_id, friend_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'};

//
// Stats
//

// Q14: Find active sessions per node by day
CREATE TABLE log_active_sessions_by_day (
	node int,
	day int,  // yyyyMMdd format
	user_id bigint,
	last_time timestamp,
	PRIMARY KEY ((node, day), user_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (user_id DESC);
 
// Q15: Find registered users by day
CREATE TABLE log_registered_users_by_day (
	day int, // yyyyMMdd format
	user_id bigint,
	created_date timestamp,
	PRIMARY KEY (day, user_id)
)
WITH COMPACTION = {'class' : 'LeveledCompactionStrategy'}
AND CLUSTERING ORDER BY (user_id DESC);
