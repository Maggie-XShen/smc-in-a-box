# **Secure Aggregation with Input Certification**
## **Objective**
The objective is ...
## **Participating Parties**
* n_c Clients
* n_s Servers
* Output Party
## **Protocol I: Semi-honest Server Model**
### **Parameters**
* Input
  + Client i has input vector of length m (i.e., x_i).
* Output
  + The output party receives xi  P(xi ) where P(.) is a robustness predicate, say L2 norm.
* Adversarial Model
  + less than n_c -1 malicious clients
  + t semi-honest servers
### **Protocol Description**
**Phase 0:** Setup

We assume secure and authenticated channels (i) from each client and all the servers and (ii) complete communication graph among the servers.

**Phase 1:** Input Sharing & Proof generation

1. Each client i has an input vector of length m which is rearranged into a matrix of dimension b x l. This input is encoded row-by-row using the packed secret sharing scheme to obtain an encoded input matrix of size b x n_s. 
2. Each column of the encoded input matrix is sent to its corresponding server i.e., column j is sent to server j for all j in [n_s].
3. The servers extend the input matrix by adding additional rows to construct the extended witness of size b’ x l and then encode into a matrix of size b’ x n_s. This is used to compute the Ligero proof .
4. The proof  and the column j of the input encoded matrix are sent to server j for all j in [n_s].
   
**Phase 2:** Input Consistency

1. Each server j performs the following checks each client i:
   - Proof verification passes
   - The hash of the received input column matches the hash received as part of the proof 
   - The random linear combination of the input column (with respect to the given randomness) matches the corresponding entry in the output of the degree test (for the input).
2. Server j initializes a set Vj with all the clients that pass all of the above three tests.
3. Also, server j sends the set Vj to a single server, server 1.
4. Upon receiving all the sets {Vj}_j \in [n_s], server 1 computes the intersection of all the sets, say V, and sends V to all the servers.

**Phase 3:** Output Reconstruction

1. Upon receiving the set V, the servers aggregate the columns received from each client in V.
2. Then the aggregated columns are sent to the output party who can then reconstruct the aggregate.



### **Building Blocks**
## **Protocol II: Malicious Server Model**
### **Parameters**
* Input
* Output
* Adversarial Model
### **Protocol Description**
### **Building Blocks**