create table events (
    id         integer generated always as identity primary key,
    created_at timestamptz not null,
    name       text not null,
    type       integer not null,
    event_date timestamptz not null
);
create table todos (
    id           integer generated always as identity primary key,
    updated_at   timestamptz not null,
    done         boolean not null,
    title        text not null,
    event_id     integer references events(id)
);