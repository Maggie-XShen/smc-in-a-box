# SCIF: Privacy-Preserving Statistics Collection with Input Validation and Full Security
This repository contains the prototype implementation of SCIF, appearing in Proceedings on Privacy Enhancing Technologies (PoPETS), 2025.

SCIF is a practical system for private statistics,  comprising clients, several secure multiparty computation servers, and output party(e.g., data analyst) written in Golang. The system guarantees that all honest clients' inputs are included in final result and all malformed inputs from malicious clients are excluded from final result. Even though a minority of malicious servers exist, they will not prevent the output party computing the final result. For more details, please refer to the [paper](https://eprint.iacr.org/2024/1821).

## Build & Run with Docker Compose
One-command local execution. This allows all parties running on the same machine.
At folder local, after preparing each party's template, compile then run
```
./local
```

## Build & Run on Cloud Provider
The deployment and execution on Google Cloud are managed through ansible. Details are described in the following repository.
https://github.com/GUSecLab/smc-in-a-box-ansible/tree/main

## Build & Run Manually
### 1. Building environment
   
   First, ensure that you have Go installed. 

   ```
   % go version
   ```
   
   Second, ensure that MySQL is installed and running with a configured username and password for connections. The default credentials in SetupDatabase() (store.go) may need to be replaced with your own.

   ```
   % mysql --version 
   ```

   If MySQL is installed, you’ll see a version number.

   ```
   % brew services list
   ```

   If MySQL is running, it will be marked as started.
   
### 2. Prepare config file and input file
   
Each client, server, and output party needs a config and input file before running. Example configs are in client_template.json, server_template.json, and outputparty_template.json. Scripts for batch generation are in each party's scripts/generator directory.

Client Config Example
```
{
    "Client_ID": "c1",
    "Token": "tk1",
    "URLs": ["http://127.0.0.1:50001/client/", 
            "http://127.0.0.1:50002/client/", 
            "http://127.0.0.1:50003/client/", 
            "http://127.0.0.1:50004/client/"], 
    "N": 4,  
    "T": 1,  
    "Q":  41543, 
    "N_secrets":10000, 
    "M": 50, 
    "N_open":240 
}
```

Key Fields:

- URLs: List of server URLs for client data submission.
- N: Total number of servers.
- T: Number of malicious servers.
- Q: Modulus.
- N_secrets: Length of the client's input vector.
- M: Number of rows in the extended witness for the Ligero ZK proof.
- N_open: Number of opened columns in the encoded extended witness.

Server Config Example
```
{
    "Server_ID": "s1",
    "Token": "stk1",
    "Cert_path":"/path to certificate", 
    "Key_path":"/path to private key",  
    "Port": "50001", 
    "Complaint_urls":[
        "http://127.0.0.1:50002/complaint/", 
        "http://127.0.0.1:50003/complaint/", 
        "http://127.0.0.1:50004/complaint/"  
    ],
    "Masked_share_urls":[
        "http://127.0.0.1:50002/maskedShare/", 
        "http://127.0.0.1:50003/maskedShare/", 
        "http://127.0.0.1:50004/maskedShare/"  
    ],
    "Share_Index": 1, 
    "N": 4, 
    "T": 1, 
    "Q":  41543, 
    "N_secrets":10000,
    "M": 50,  
    "N_open":240
}
```

Key Fields:

- Cert_path: Server certificate location (required for TLS).
- Key_path: Server private key location (required for TLS).
- Port: Port for client and server connections.
- Complaint_urls: List of server URLs for submitting complaints.
- Masked_share_urls: List of server URLs for submitting masked shares.
- Share_Index: Server ID index (e.g., 1 for server s1).

Output Party Config Example 
```
{
    "OutputParty_ID": "op1",
    "Cert_path":"/path to certificate",
    "Key_path":"/path to private key",
    "Port": "60000",
    "N": 4,
    "T": 1,
    "N_secrets": 10000,
    "Q": 41543
}
```
Key Fields:

- Cert_path: Output party certificate location (required for TLS).
- Key_path: Output party private key location (required for TLS).
- Port: Port for server connections.

Client Input Example
```
[
   {"Exp_ID":"exp1","Secrets":[0,1]},
   {"Exp_ID":"exp2","Secrets":[1,1]}
]
```

Server Input Example
```
[
   {"Exp_ID":"exp1",
   "ClientShareDue":"2025-03-11 18:13:57.188395 +0000 UTC",  
   "ComplaintDue":"2025-03-11 18:15:57.188395 +0000 UTC", 
   "ShareBroadcastDue":"2025-03-11 18:17:57.188395 +0000 UTC", 
   "Owner":"http://127.0.0.1:60000/serverShare/"}
]
```

Output Party Input Example
```
[
   {"Exp_ID":"exp1",
   "ClientShareDue":"2025-03-11 18:13:57.188395 +0000 UTC",
   "ServerShareDue":"2025-03-11 18:19:57.188395 +0000 UTC"} 
]
```

Key Fields:

- Secretes: Client input vector, with each bit representing an attribute.
- ClientShareDue: Deadline for clients to submit shares and proofs.
- ComplaintDue: Deadline for servers to submit complaints.
- ShareBroadcastDue: Deadline for servers to share masked data.
- ServerShareDue: Deadline for servers to submit aggregated shares to the output party.
- Owner: URL of the output party for servers to submit aggregated shares.

### 3. Run the software
To start up a server, in the server/cmd folder, compile and run:
```
./cmd -confpath="path_to_server_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
```

To start up an output party, in the outputparty/cmd folder, compile and run:
```
./cmd -confpath="path_to_output_party_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
```

To start up a client, in the folder client/cmd, compile then run
```
./cmd -confpath=“path_to_client_config_file” -inputpath=“path_to_input_file” -logpath="path_to_log_folder"
``` 

Parameter Descriptions:
- For server and output party: use -mode="http" to disable TLS; the default enables it (which requires setup of certificate).
- For server: use -n_client_mal=num_of_malicious_client to enable malicious clients; the default assumes all are honest.
- For client: use -mode=honest to run client without malicious behavior. Default setting assumes client could act maliciously.
   
 **Note:** Servers and the output party must start before clients.


## Citation
If you find this work useful, please cite it as follows:




  
