# Postgres and Valkey as day-one infrastructure

The template ships with Postgres as its database and Valkey as its session store from the first commit, both run locally through Docker. A file-based database and in-database sessions would make a fresh clone lighter, but every project built from this template is expected to outgrow them, and swapping a database or session store after features accumulate costs far more than running two containers from the start.

## Considered Options

- **SQLite, sessions in the database**: zero services to run, single-file deploys; rejected because migrating real projects off it later is the expensive path this template exists to avoid.
- **Postgres, sessions in Postgres**: one fewer service; rejected because sessions are high-churn, short-lived data that a key-value store handles natively with expiry, and the same store also serves later needs such as rate limiting and realtime fan-out.
- **Redis instead of Valkey**: the original pick; Valkey is BSD-licensed where Redis 8 is tri-licensed with AGPL, and it speaks the same protocol, so Redis clients work unchanged.
