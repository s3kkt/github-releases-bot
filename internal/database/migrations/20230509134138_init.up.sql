-- Table with list of repos
CREATE TABLE IF NOT EXISTS repos (
    name           varchar,
    deleted        boolean DEFAULT false,
    chat_id        bigint,
    PRIMARY KEY (name, chat_id)
);
-- Table with releases for repos from list
CREATE TABLE IF NOT EXISTS releases (
    release_url    varchar,
    author         varchar,
    tag_name       varchar,
    release_name   varchar,
    target_branch  varchar,
    is_draft       bool,
    is_prerelease  bool,
    created_at     timestamp,
    published_at   timestamp,
    last_check     timestamp,
    tarball_url    varchar,
    zipball_url    varchar,
    release_text   varchar,
    repo_name      varchar,
    PRIMARY KEY (repo_name, tag_name)
);