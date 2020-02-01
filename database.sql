create table User
(
    id                 integer not null
        constraint User_pk
            primary key autoincrement,
    email              text,
    password_encrypted text,
    name               text,
    role               text
);

create unique index User_id_uindex
    on User (id);

create table File
(
    id                 integer not null
        constraint file_pk
            primary key autoincrement,
    name               text,
    version            text,
    mirror_url         text,
    mirror_file_status text,
    local_file_status  text,
    download_time      integer default 0,
    uuid               text,
    mirror_carrier     text
);

create unique index file_id_uindex
    on File (id);



create table Task
(
    id                integer not null
        constraint Task_pk
            primary key autoincrement,
    file_id           integer,
    download_progress real,
    user_id           integer,
    origin_url        text,
    final_url         text,
    status            text
);

create unique index Task_id_uindex
    on Task (id);
