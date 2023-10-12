# protocol version: dropout clients, semi-honest server
This version introduces server-to-server communication, which facilitates the identification of the overlap in their respective client spaces. When the intersection of their client spaces is empty, the output party cannot comlete the computation.

## Basic Features
+ Client
  + Client can connect to several servers.
  + Client can generate shares of a secret.
  + Client can send secret share to a server.
+ Server
  + Server can connect to the output party.
  + Server can write experiment information from output party to database.
  + Server can write secret share of each client to database.
  + Server can aggregate secret share from each client and send result to output party. 
  + Server can connect to the other servers.
  + Server can write other servers' client spaces to database.
  + Server can find intersection of client spaces.
+ Output Party
  + Output party can connect to the server.
  + Output party can send experiment information to the server.
  + Output party can write aggregated secret shares from server to database.
  + Output party can reveal the sum of clients’ secrets without learning anything about client’s secret.

## Protocol
Phase 1: Output party sends experiment information to servers.

Phase 2: Clients register experiments.

Phase 3: Clients send secret's shares to servers.

Phase 4: Servers send their client space to each other after due of experiment and identify the overlap

Phase 5: Servers compute sum of shares of common clients and send result to output party.

Phase 6: Output party reveals sum of secrets.

## Implementation (need to modify)
### Client
+ Configuration
  + client_id, token(e.g. password), servers_urls, N (number of servers),T, K(number of secrets), Q(prime number), experiment_id, secret

### Server
+ Configuration
  + server_id, token, port(listen client request), sumshare_index, output party url
+ Database
  + Experiment Table: experiment_id, due, completed
  + Client Table: experiment_id, client_id, share_index, share_value
  + Client Registry Table: client_id, token, experiment_id
  
### Output Party
+ Configuration
  + outputparty_id, port(listen server request), N, T, K, Q, servers_urls
+ Database
  + Experiment Table: experiment_id, due, completed
  + Server Table: experiment_id, server_id, sumshare_value, sumshare_index