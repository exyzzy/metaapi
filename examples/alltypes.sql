create table todos (
  id           integer generated always as identity primary key,
  updated_at   timestamptz,
  done         boolean,
  title        text
);

create table allbools (
  id         serial primary key,
  abool      boolean,
  abool2     bool
);

create table allchars (
  id         serial primary key,
  achar      char(16),
  avarchar   varchar(16),
  atext      text
);

create table allints (
  id         serial primary key,
  asmallint  smallint,
  aint       int,
  aint2      integer,
  asmallser  smallserial,
  aser       serial,
  abigser    bigserial
);

create table allfloats (
  id         serial primary key,
  afloat     float(53),
  areal      real,
  afloat8    float8,
  adecimal   decimal,
  anumeric   numeric,
  anumeric2  numeric(36,18),
  adouble    double precision
);

create table alltimes (
  id         serial primary key,
  adate      date,
  atime      time,
  ats        timestamp,
  atsz       timestamptz,
  ainterval  interval
);

create table alljsons (
  id         serial primary key,
  ajson      json,
  asjonb     jsonb
);

create table uuids (
  id         serial primary key,
  auuid      uuid
);