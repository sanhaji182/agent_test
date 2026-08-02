-- 014_llm_profiles: multi-provider LLM configuration profiles.
-- Up
CREATE TABLE IF NOT EXISTS llm_profiles (
    id          VARCHAR(64)  PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    provider    VARCHAR(64)  NOT NULL DEFAULT '',
    base_url    TEXT         NOT NULL DEFAULT '',
    api_key     TEXT         NOT NULL DEFAULT '',
    model       VARCHAR(255) NOT NULL DEFAULT '',
    temperature VARCHAR(16)  NOT NULL DEFAULT '',
    max_tokens  VARCHAR(16)  NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_llm_profiles_active ON llm_profiles (is_active);

-- Down
-- DROP TABLE IF EXISTS llm_profiles;
