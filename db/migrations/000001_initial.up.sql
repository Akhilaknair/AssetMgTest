begin;
create extension if not exists "pgcrypto";
create type asset_type as enum (
    'laptop',
    'keyboard',
    'mouse',
    'mobile'
    );

create type asset_status as enum (
    'available',
    'assigned',
    'in_service',
    'waiting_for_repair',
    'damaged'
    );

create type user_role as enum (
    'admin',
    'employee',
    'asset-manager',
    'project-manager'
    );

create type user_type as enum(
    'full-time',
    'intern',
     'freelancer'
    );

create type owner_type as enum (
    'client',
    'company'
    );

create table if not exists users
(
    id            uuid primary key default gen_random_uuid(),
    name          text not null,
    email         text not null,
    role          user_role        default 'employee',
    phone_no      text not null,
    joining_date     date not null,
    password text not null,
    user_type     user_type not null,
    created_at    timestamptz        default current_timestamp,
    archived_at   timestamptz
    );

create unique index idx_unique_email on users (email) where archived_at is NULL;


create table assets
(
    id          uuid primary key default gen_random_uuid(),
    brand       text        not null,
    model       text        not null,
    serial_no   text        not null,
    asset_type  asset_type  not null,
    status      asset_status     default 'available',
    warranty_start_date timestamptz not null,
    warranty_expiry_date  timestamptz not null,
    owner       owner_type  not null,
    created_at  timestamptz  default current_timestamp,
    archived_at timestamptz,
    updated_at timestamptz,

    constraint check_warranty_date
        check(warranty_start_date<=warranty_expiry_date)

);

create unique index idx_unique_asset_serial on assets (serial_no) where archived_at is null;

create table asset_history
(
    id            uuid primary key default gen_random_uuid(),
    asset_id      uuid references assets (id) not null,
    assigned_to   uuid references users (id) not null,
    assigned_by  uuid references  users(id) not null,
    assigned_on   timestamptz not null,
    service_start timestamptz,
    service_end timestamptz,
    returned_on   timestamptz,

    constraint check_service_date
        check((service_start is null and service_end is null)
             or (service_end >= service_start)
             and (returned_on is null or returned_on >= assigned_on))
);

create unique index idx_active_assignment on asset_history(asset_id) where returned_on is null;

create table if not exists user_session
(
    id          uuid primary key default gen_random_uuid(),
    user_id      uuid references users (id) NOT NULL,
    created_at  timestamptz  default current_timestamp,
    archived_at timestamptz
    );

create table laptop
(
    id uuid primary key default gen_random_uuid(),
    asset_id uuid unique references assets (id) not null,
    processor text,
    ram       text,
    storage   text,
    os        text,
    charger   text,
    password  text not null
);

create type wire_type as enum ('wired', 'wireless');

create table keyboard
(
    id uuid primary key default gen_random_uuid(),
    asset_id uuid unique references assets (id) not null,
    layout text,
    connectivity wire_type
);

create table mouse
(
    id uuid primary key default gen_random_uuid(),
    asset_id uuid unique references assets (id) not null,
    dpi int,
    connectivity wire_type
);

create table mobile
(
    id uuid primary key default gen_random_uuid(),
    asset_id uuid unique references assets (id) not  null,
    os       text not null,
    ram      text not null,
    storage  text not null,
    charger  text,
    password text not null
);

commit;