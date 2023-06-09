create table if not exists public.tc_devices
(
    id                  integer generated by default as identity
        primary key,
    name                varchar(128) not null,
    uniqueid            varchar(128) not null
        unique,
    lastupdate          timestamp,
    positionid          integer,
    groupid             integer,
    attributes          varchar(4000),
    phone               varchar(128),
    model               varchar(128),
    contact             varchar(512),
    category            varchar(128),
    disabled            boolean          default false,
    status              char(8),
    "geofenceIds"       varchar(128),
    expirationtime      timestamp,
    motionstate         boolean          default false,
    motiontime          timestamp,
    motiondistance      double precision default 0,
    overspeedstate      boolean          default false,
    overspeedtime       timestamp,
    overspeedgeofenceid integer          default 0,
    motionstreak        boolean          default false
);

create index if not exists idx_devices_uniqueid
    on public.tc_devices (uniqueid);

create table public.tc_positions
(
    id         bigint generated by default as identity,
    protocol   varchar(128),
    deviceid   integer                                                                           not null,
    servertime timestamp        default '2023-02-16 10:19:13.70023'::timestamp without time zone not null,
    devicetime timestamp                                                                         not null,
    fixtime    timestamp                                                                         not null,
    valid      boolean                                                                           not null,
    latitude   double precision                                                                  not null,
    longitude  double precision                                                                  not null,
    altitude   double precision                                                                  not null,
    speed      double precision                                                                  not null,
    course     double precision                                                                  not null,
    address    varchar(512),
    attributes varchar(4000),
    accuracy   double precision default 0                                                        not null,
    network    varchar(4000)
);

create index position_deviceid_fixtime
    on public.tc_positions (deviceid, fixtime);

create index position_servertime
    on public.tc_positions (servertime);
