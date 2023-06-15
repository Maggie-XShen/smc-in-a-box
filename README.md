# Secure Aggregation with Input Certification

| Parameter              | Description                                                                              |
|------------------------|------------------------------------------------------------------------------------------|
| Participating Parties  | nc clients $\{u_1, \ldots, u_{nc}\}$, ns servers $\{s_1, \ldots, s_{ns}\}$, and output party |
| Input                  | Client $u_i$ has input vector $x_i$ of length $m$                                          |
| Output                 | The output party receives $\sum x_i \cdot P(x_i)$ where $P(.)$ is a robustness predicate, say L2 norm. |
| Adversarial Model      | $< nc - 1$ malicious client, $ts$ semi-honest servers                                         |


## Protocol Description

### Phase 0: Setup
We assume secure and authenticated channels from each client and all the servers and a complete communication graph among the servers.

### Phase 1: Input Sharing & Proof Generation
1. Each client $c_i$ generates the shares by invoking $Share(x_i)$ which outputs $(sh_{i,1}, \ldots, sh_{i,ns})$ where $sh_{i,j}$ is the server $s_j$’s share.
2. Proof generation: Each client generates the proof as per the Ligero proof system. Specifically, the $ProofGen(x_i, sh_1, \ldots, sh_{ns})$ outputs the proof $\pi_i = (\pi_{i,0}, \ldots, \pi_{i,ns})$ where $\pi_{i,0}$ is given to all servers and $\pi_{i,j}$ is given to server $s_j$.
3. Each client $u_i$ sends the following to server $s_j$ for all $j$ in $[ns]$:
	- Share $sh_{i,j}$
	- Proof $(\pi_{i,0}, \pi_{i,j})$
	- $h_{i,j} = Hash(\pi_{i,0})$

### Phase 2: Client Elimination
4. Each server $s_j$ locally performs the following checks for each client $u_i$ and sets a bit $happy_i = 0$ if any of them fail, otherwise $happy_i = 1$:
	- $h_{i,j}$ is valid, i.e., check if $h_{i,j} = Hash(\pi_{i,0})$
	- $Proof.Verify(\pi_{i,0}, \pi_{i,j}, sh_{i,j})$ outputs 1
		<!--- - (Ligero Hash Check) The hash of the received input column matches the hash received as part of the proof $\pi_{i,0}$
		- (Ligero Degree test) The random linear combination of the input column (with respect to the given randomness) matches the corresponding entry in the output of the degree test (for the input).
		- (Ligero Linear test) Checks that the shares $sh_{i,j}$ are a linear combination of the inputs used in the proof $\pi_{i,0}$ -->

5. Server $s_j$ initializes a set $V_j$ with all the clients that pass all of the above checks, i.e., clients with $happy_i = 1$.
6. Each server $s_j$ sends the tupleS \{ $(i, h_{i,j})$ \} $_{u_i \in V_j}$ to all the servers.
7. The servers compute a set $V$ such that client $u_i$ is included in $V$ if:
	<!--- - All servers sent messages of the form $(i, *)$ -->
	- All servers sent the same hash with respect to client $i$, i.e., $h_{i,k} = h_{i,l} \neq \bot$ for all $k, l \in [ns]$

### Phase 3: Output Reconstruction
8. Each server $s_j$ aggregates the shares it received from all clients in $V$ as follows: $O_j = \sum_{u_i\in V} sh_{i,j}$ and send $O_j$ to the output party.
9. The output party reconstructs $(O_1, \ldots, O_{ns})$ to obtain the aggregate.


## Implementation Details

### Executing Program
+ To test client end, go to the folder SMC/cmd/client and run command 

	go run main.go -cid="c1"
+ To test server end, go to the folder SMC/cmd/server and run command 
  
  go run main.go -port=":8080" -sid="s1"

### Current Status
+ Client end could splite single secret and send to servers. Secret, servers' urls and parameters for packed secret sharing can be read from config file.
+ Server end could receive share from different client. 

### To Do
+ Test aggregating shares and reconstruct sum of secrets.
+ Let server listen to client, aggregate shares and send to output party concurrently.
+ Change go.mod to import packed secret sharing from github 
  
