# Postgres and Redis as day-one infrastructure

The template ships with Postgres as its database and Redis as its session store from the first commit, both run locally through Docker. A file-based database and in-database sessions would make a fresh clone lighter, but every project built from this template is expected to outgrow them, and swapping a database or session store after features accumulate costs far more than running two containers from the start.

## Considered Options

- **SQLite, sessions in the database**: zero services to run, single-file deploys; rejected because migrating real projects off it later is the expensive path this template exists to avoid.
- **Postgres, sessions in Postgres**: one fewer service; rejected because sessions are high-churn, short-lived data that Redis handles natively with expiry, and Redis also serves later needs such as rate limiting and realtime fan-out.
