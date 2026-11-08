CREATE TABLE IF NOT EXISTS hosts (
    host       text PRIMARY KEY,
    first_seen timestamptz NOT NULL,
    last_seen  timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS snapshots (
    id               bigserial PRIMARY KEY,
    host             text NOT NULL REFERENCES hosts(host),
    ts               timestamptz NOT NULL,
    cpu_percent      double precision NOT NULL,
    mem_total_kb     bigint NOT NULL,
    mem_free_kb      bigint NOT NULL,
    mem_available_kb bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS snapshots_host_ts ON snapshots (host, ts DESC);

CREATE TABLE IF NOT EXISTS process_samples (
    id          bigserial PRIMARY KEY,
    snapshot_id bigint NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
    pid         integer NOT NULL,
    name        text NOT NULL,
    state       text NOT NULL,
    threads     bigint NOT NULL,
    rss_kb      bigint NOT NULL,
    cpu_ticks   bigint NOT NULL,
    cpu_percent double precision NOT NULL
);
CREATE INDEX IF NOT EXISTS process_samples_snapshot ON process_samples (snapshot_id);