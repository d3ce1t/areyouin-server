//
// Edit /usr/bin/cqlsh.py and add csv.field_size_limit(sys.maxsize)
//

COPY user_email_credentials FROM 'user_email_credentials.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY user_facebook_credentials FROM 'user_facebook_credentials.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY user_account FROM 'user_account.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY friends_by_user FROM 'friends_by_user.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY groups_by_user FROM 'groups_by_user.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY friends_by_group FROM 'friends_by_group.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY events_by_user FROM 'events_by_user.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY event FROM 'event.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY thumbnails FROM 'thumbnails.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY user_access_token FROM 'user_access_token.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY friend_requests_received FROM 'friend_requests_received.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY friend_requests_sent FROM 'friend_requests_sent.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY log_active_sessions_by_day FROM 'log_active_sessions_by_day.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;

COPY log_registered_users_by_day FROM 'log_registered_users_by_day.csv'
WITH MAXBATCHSIZE=1 AND MINBATCHSIZE=1;
