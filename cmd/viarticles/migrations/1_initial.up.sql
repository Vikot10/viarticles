create table article (
    id serial primary key,
    title text not null,
    body text not null,
    category_id integer not null,
    created_at timestamp default now(),
    updated_at timestamp,
)

create table article_category (
    id serial primary key,
    category_id integer not null,
    article_id integer not null,
)

create table category (
    id serial primary key,
    title text not null,
    created_at timestamp default now(),
    updated_at timestamp,
)