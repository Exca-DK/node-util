CREATE TABLE "nodes" (
  "id" varchar PRIMARY KEY,
  "added_at" timestamptz NOT NULL DEFAULT (now()),
  "seen_at" timestamptz NOT NULL DEFAULT (now()),
  "queries" int NOT NULL DEFAULT 0,
  "ips" varchar[]
);

CREATE TABLE "enr_records" (
  "id" varchar PRIMARY KEY,
  "fields" varchar[],
  "added_at" timestamptz NOT NULL DEFAULT (now()),
  "edited_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "rlp_records" (
  "id" varchar PRIMARY KEY,
  "raw_name" varchar NOT NULL,
  "protocol_version" int NOT NULL,
  "caps" varchar[],
  "added_at" timestamptz NOT NULL DEFAULT (now()),
  "edited_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "scan_records" (
  "id" bigint PRIMARY KEY,
  "node" varchar NOT NULL,
  "ip" varchar NOT NULL,
  "ports" int[] NOT NULL,
  "scanned_at" timestamptz NOT NULL DEFAULT (now()),
  "edited_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "rlp_records" ("protocol_version");

CREATE INDEX ON "scan_records" ("node");

CREATE INDEX ON "scan_records" ("ip");

CREATE UNIQUE INDEX ON "scan_records" ("node", "ip");

ALTER TABLE "enr_records" ADD FOREIGN KEY ("id") REFERENCES "nodes" ("id");

ALTER TABLE "rlp_records" ADD FOREIGN KEY ("id") REFERENCES "nodes" ("id");
