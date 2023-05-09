ALTER TABLE repos DROP CONSTRAINT repos_pkey;

ALTER TABLE repos DROP COLUMN id;

ALTER TABLE repos ADD PRIMARY KEY (name);

ALTER TABLE repos ADD COLUMN from_config bool DEFAULT false;

ALTER TABLE repos ADD COLUMN deleted bool DEFAULT false;

ALTER TABLE repos ADD COLUMN chat_id int default 0;

ALTER TABLE releases DROP CONSTRAINT releases_repo_id_tag_name_key;

ALTER TABLE releases DROP COLUMN repo_id;

ALTER TABLE releases ADD COLUMN repo_name varchar DEFAULT Null;

ALTER TABLE releases ADD CONSTRAINT repo_name_tag_name UNIQUE (repo_name, tag_name);