
create table todos (
  id           integer generated always as identity primary key,
  updated_at   timestamptz,
  done         boolean,
  title        text
);

