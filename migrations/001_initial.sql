CREATE TABLE tenants (id text primary key, name text not null, quota integer not null, created_at timestamptz not null);
CREATE TABLE collections (id text primary key, tenant_id text not null, name text not null, mappings jsonb not null, mapping_version integer not null);
