CREATE TABLE IF NOT EXISTS chats (
    chat_id bigint,
    username varchar unique,
    first_name varchar default Null,
    last_name varchar default Null,
    type varchar,
    is_bot bool,
    last_activity bigint,
    PRIMARY KEY (username, chat_id)
);

CREATE EXTENSION pgcrypto;

CREATE TABLE IF NOT EXISTS chat_settings (
    chat_id bigint unique,
    update_interval varchar,
    github_token varchar
);
