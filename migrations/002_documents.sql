CREATE TABLE documents (id text primary key, tenant_id text not null, collection_id text not null, version bigint not null, data jsonb not null, deleted boolean not null default false, updated_at timestamptz not null);
CREATE TABLE snapshots (id text primary key, collection_id text, checksum text not null, state text not null, created_at timestamptz not null);
