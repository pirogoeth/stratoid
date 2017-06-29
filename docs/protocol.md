Stratum-ish Protocol
====================

CryptoNight mining protocol is a slightly modified version of the Stratum protocol used for [Bitcoin](https://en.bitcoin.it/wiki/Stratum_mining_protocol) and [Ethereum](https://github.com/nicehash/Specifications/blob/master/EthereumStratum_NiceHash_v1.0.0.txt).

Note that this documentation was written based on exchanges between [cpuminer-multi](https://github.com/tpruvot/cpuminer-multi), [Minergate](https://minergate.com) pool, and the [cryptonote-universal-pool](https://github.com/xdn-project/cryptonote-universal-pool/blob/master/lib/pool.js).

----

## Message Structure

### Requests (from Miner)

All requests from the miner to a pool server *should* contain the following fields:

- `id (int)` :: Command ID - must be provided for miner to understand responses as pool may not process requests in FIFO order
- `method (string)` :: RPC method to call in pool
- `params (object)` :: `string`->`value` mapping of parameters to pass to RPC method


### Responses (from Pool)

Responses from the pool *should* contain the following fields:

- `jsonrpc (string)` :: Version of JSON-RPC protocol the pool server is "speaking"
- `result (object)` :: Result of RPC execution. May also be `null`.

Responses from the pool *may* contain the following fields:

- `id (int)` :: If provided, indicates the request that directly triggered this response
- `method (string)` :: RPC method to call on miner
- `params (object)` :: `string`->`value` mapping of parameters to pass to RPC method
- `error (object)` :: `string`->`value` mapping describing the error that occurred

----

## Miner RPC Commands

*This list created from [`stratum_handle_method`](https://github.com/tpruvot/cpuminer-multi/blob/7495361e34bb11e0c3e2c778312281071208eb55/util.c#L2058) in `cpuminer-multi`*

### `job`

Submits a job to the miner to begin execution.

Parameters:

|  Key  |  Type  |  Description  |
|---|---|---|
|  blob  |  string  |  Hex-encoded binary blob  |
|  job_id  |  string  |  Job UUID  |
|  target  |  string  |  Hex-encoded binary nonce?  |
|  time_to_live  |  int  |  Lifetime length of job  |

### `mining.notify`


### `mining.ping`


### `mining.set_difficulty`


### `mining.set_extranonce`


### `client.reconnect`


### `client.get_algo`


### `client.get_stats`


### `client.get_version`


### `client.show_message`


----

## Pool RPC Commands

*This list created from [`cryptonote-universal-pool`](https://github.com/xdn-project/cryptonote-universal-pool).*

### `login`

Authenticate the remote client with the pool server.

Parameters:

|  Key  |  Type  |  Description  |
|---|---|---|
|  login  |  string  |  Miner login username  |
|  pass  |  string  |  Miner login password  |

Returns:

|  Key  |  Type  |  Description  |
|---|---|---|
|  result.id  |  string  |  Miner auth ID  |
|  result.job  |  dictionary  |  Job data  |
|  result.status  |  string  |  Login status  |
|  error  |  dictionary  |  Mapping of `string`->`value` describing error, otherwise null  |

### `getjob`

Retrieves a new job for the miner to process.

Parameters: *none*.

Returns:

|  Key  |  Type  |  Description  |
|---|---|---|
|  result.blob  |  string  |  Hex-encoded binary blob  |
|  result.job_id  |  string  |  Job UUID  |
|  result.target  |  string  |  Hex-encoded binary nonce?  |
|  error  |  dictionary  |  Mapping of `string`->`value` describing error, otherwise null  |

### `submit`

Submits a proof of work to the pool for verification.

Parameters:

|  Key  |  Type  |  Description  |
|---|---|---|
|  result.id  |  string  |  Miner ID  |
|  result.job_id  |  string  |  Job UUID  |
|  result.nonce  |  string  |  Hex-encoded nonce data  |
|  result.result  |  string  |  Hex-encoded result data  |
|  error  |  dictionary  |  Mapping of `string`->`value` describing error, otherwise null  |

Returns:

|  Key  |  Type  |  Description  |
|---|---|---|
|  result.status  |  string  |  Submission status  |
|  error  |  dictionary  |  Mapping of `string`->`value` describing error, otherwise null  |