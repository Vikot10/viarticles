create table article (
    id serial primary key,
    title nvarchar(255) not null,
    body text default '',
    url nvarchar(255) not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);

create table article_category (
    category_id integer not null,
    article_id integer not null
);

create table category (
    id serial primary key,
    title text not null,
    created_at timestamp default now(),
    updated_at timestamp,
    deleted_at timestamp
);