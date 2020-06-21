create table todos (
  id           integer generated always as identity primary key,
  updated_at   timestamptz not null,
  done         boolean not null,
  title        text not null
);

create table allbools (
  id         serial primary key,
  abool      boolean not null,
  abool2     bool not null
);

create table allchars (
  id         serial primary key,
  achar      char(16) not null,
  avarchar   varchar(16) not null,
  atext      text not null
);

create table allints (
  id         serial primary key,
  asmallint  smallint not null,
  aint       int not null,
  aint2      integer not null,
  asmallser  smallserial not null,
  aser       serial not null,
  abigser    bigserial not null
);

create table allfloats (
  id         serial primary key,
  afloat     float(53) not null,
  areal      real not null,
  afloat8    float8 not null,
  adecimal   decimal not null,
  anumeric   numeric not null,
  anumeric2  numeric(36,18) not null,
  adouble    double precision not null
);

create table alltimes (
  id         serial primary key,
  adate      date not null,
  atime      time not null,
  ats        timestamp not null,
  atsz       timestamptz not null,
  ainterval  interval not null
);

create table alljsons (
  id         serial primary key,
  ajson      json not null,
  asjonb     jsonb not null
);

create table uuids (
  id         serial primary key,
  auuid      uuid not null
);

create table todosnull (
  id           integer generated always as identity primary key,
  updated_at   timestamptz,
  done         boolean,
  title        text
);

create table allboolsnull (
  id         serial primary key,
  abool      boolean,
  abool2     bool
);

create table allcharsnull (
  id         serial primary key,
  achar      char(16),
  avarchar   varchar(16),
  atext      text
);

create table allintsnull (
  id         serial primary key,
  asmallint  smallint,
  aint       int,
  aint2      integer,
  asmallser  smallserial,
  aser       serial,
  abigser    bigserial
);

create table allfloatsnullish (
  id         serial primary key,
  afloat     float(53),
  areal      real not null,
  afloat8    float8,
  adecimal   decimal,
  anumeric   numeric,
  anumeric2  numeric(36,18),
  adouble    double precision
);

create table alltimesnull (
  id         serial primary key,
  adate      date,
  atime      time,
  ats        timestamp,
  atsz       timestamptz,
  ainterval  interval
);

create table alljsonsnull (
  id         serial primary key,
  ajson      json,
  asjonb     jsonb
);

create table uuidsnull (
  id         serial primary key,
  auuid      uuid
);