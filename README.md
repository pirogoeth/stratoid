stratoid
========

An intelligent Stratum (and Stratum-ish) proxy for cryptocurrency mining.

----

## Goals

- Build a *fast* proxy system supporting both Stratum and Stratum-derived protocols (cryptonight, etc..)
- Be able to forward to pools with as little of overhead as possible
- Have a proxy control plane which allows management of your workers:
    - Retargeting specific workers to a different pool
    - Drain of all workers from the pool once they complete their current jobs
- Provide detailed stats on hash rate, temperature, etc, for workers via an API.
- Provide simple management of the `stratoid` server via `stratctl` and/or REST API.

----

## Status

`stratoid` (and `stratctl`) are both in the early stages of development. As more work is accomplished, this
portion of the readme will be updated accordingly.