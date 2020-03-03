create table owners (
  id           integer generated always as identity primary key,
  name         text
);

create table todos (
  id           integer generated always as identity primary key,
  updated_at   timestamptz,
  done         boolean,
  title        text,
  owner        integer references owners(id)
);

