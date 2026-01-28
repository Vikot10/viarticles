-- Источники статей
CREATE TYPE article_source AS ENUM ('manual', 'vk', 'telegram', 'habr');

CREATE TABLE category (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_category_title ON category(title);
CREATE INDEX idx_category_deleted_at ON category(deleted_at);

CREATE TABLE article (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    body TEXT DEFAULT '',
    url VARCHAR(2048) NOT NULL,
    source article_source NOT NULL DEFAULT 'manual',
    source_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT unique_source_id UNIQUE (source, source_id)
);

CREATE INDEX idx_article_url ON article(url);
CREATE INDEX idx_article_source ON article(source);
CREATE INDEX idx_article_deleted_at ON article(deleted_at);
CREATE INDEX idx_article_created_at ON article(created_at DESC);

CREATE TABLE article_category (
    article_id INTEGER NOT NULL REFERENCES article(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES category(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, category_id)
);

CREATE INDEX idx_article_category_article ON article_category(article_id);
CREATE INDEX idx_article_category_category ON article_category(category_id);