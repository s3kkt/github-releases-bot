-- Table with list of repos
CREATE TABLE IF NOT EXISTS repos (
            id             SERIAL PRIMARY KEY,
            name           varchar UNIQUE
);
-- Table with releases for repos from list
CREATE TABLE IF NOT EXISTS releases (
            repo_id        int,
            release_url    varchar,
            author         varchar,
            tag_name       varchar,
            release_name   varchar,
            target_branch  varchar,
            is_draft       boolean,
            is_prerelease  boolean,
            created_at     timestamp,
            published_at   timestamp,
            last_check     timestamp,
            tarball_url    varchar,
            zipball_url    varchar,
            release_text   varchar,
            UNIQUE (repo_id, tag_name)
);