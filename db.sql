-- CREATE DATABASE streampets;
-- CREATE SCHEMA public AUTHORIZATION pg_database_owner;

CREATE TABLE channels (
	channel_id text NOT NULL,
	channel_name text NULL,
	overlay_id text NULL,
	CONSTRAINT channels_pk PRIMARY KEY (channel_id)
);

CREATE TABLE items (
	item_id uuid NOT NULL,
	"name" varchar NOT NULL,
	rarity varchar NOT NULL,
	image varchar NOT NULL,
	prev_img varchar NOT NULL,
	CONSTRAINT items_pk PRIMARY KEY (item_id)
);

CREATE TABLE users (
	user_id varchar NOT NULL,
	username varchar NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (user_id)
);

CREATE TABLE channel_items (
	channel_id varchar NOT NULL,
	item_id uuid NOT NULL,
	CONSTRAINT channelitems_pk PRIMARY KEY (channel_id, item_id),
	CONSTRAINT channelitems_unique UNIQUE (item_id),
	CONSTRAINT channelitems_channels_fk FOREIGN KEY (channel_id) REFERENCES channels(channelid),
	CONSTRAINT channelitems_items_fk FOREIGN KEY (item_id) REFERENCES items(item_id)
);

CREATE TABLE default_channel_items (
	channel_id varchar NOT NULL,
	item_id uuid NOT NULL,
	CONSTRAINT defaultchannelitems_pk PRIMARY KEY (channel_id),
	CONSTRAINT defaultchannelitems_channelitems_fk FOREIGN KEY (channel_id,item_id) REFERENCES channel_items(channel_id,item_id)
);

CREATE TABLE owned_items (
	user_id varchar NOT NULL,
	transaction_id varchar NOT NULL,
	item_id uuid NOT NULL,
	channel_id varchar NOT NULL,
	CONSTRAINT owneditems_pk PRIMARY KEY (user_id, item_id, channel_id),
	CONSTRAINT owneditems_unique UNIQUE (transaction_id),
	CONSTRAINT owneditems_channelitems_fk FOREIGN KEY (channel_id,item_id) REFERENCES channel_items(channel_id,item_id),
	CONSTRAINT owneditems_users_fk FOREIGN KEY (user_id) REFERENCES users(user_id)
);
CREATE INDEX owneditems_userid_idx ON public.owned_items USING btree (user_id, channel_id);

CREATE TABLE selected_items (
	user_id varchar NOT NULL,
	channel_id varchar NOT NULL,
	item_id uuid NOT NULL,
	CONSTRAINT selecteditems_pk PRIMARY KEY (user_id, channel_id),
	CONSTRAINT selecteditems_channelitems_fk FOREIGN KEY (channel_id,item_id) REFERENCES channel_items(channel_id,item_id),
	CONSTRAINT selecteditems_users_fk FOREIGN KEY (user_id) REFERENCES users(user_id)
);