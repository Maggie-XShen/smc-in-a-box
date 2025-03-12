# SCIF: Privacy-Preserving Statistics Collection with Input Validation and Full Security
This repository contains the SCIF prototype, built in Golang. SCIF is a practical private statistics system, consisting of clients, secure multiparty computation servers, and an output party (e.g., data analyst). It ensures that honest clients' inputs are included in the final result (e.g., sum), while excluding malicious inputs. Even with a minority of malicious servers, the output party can still compute the final result. For more details, please [refer to our paper](#citation).

## Build & Run with Docker Compose
To start a cluster of servers, an output party and a cluster of clients, go to the local folder and run:
```
$ ./local
```

## Build & Run on Cloud Provider
The deployment and execution on Google Cloud are managed through ansible. Details are described in the following repository.
https://github.com/GUSecLab/smc-in-a-box-ansible/tree/main

## Build & Run Manually
### 1. Building environment
   
   First, ensure that you have Go installed. 

   ```
   $ go version
   ```
   
   Second, ensure that MySQL is installed and running with a configured username and password for connections. The default credentials in SetupDatabase() (store.go) may need to be replaced with your own.

   ```
   $ mysql --version 
   ```

   If MySQL is installed, you’ll see a version number.

   ```
   $ brew services list
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
- N, T, Q, N_secrets are same for server, client and output party.

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
- N, T, Q, N_secrets are same for server, client and output party.

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
Before starting any party, in the smc-in-a-box folder, run the following command line to ensure that all dependencies are properly fetched
```
$ go mod tidy
```

To start a server, in the server/cmd folder, compile and run:
```
$ ./cmd -confpath="path_to_server_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
```

To start an output party, in the outputparty/cmd folder, compile and run:
```
$ ./cmd -confpath="path_to_output_party_config_file" -inputpath="path_to_experiments_file" -logpath="path_to_log_folder" -n_client=num_of_clients
```

To start a client, in the folder client/cmd, compile then run
```
$ ./cmd -confpath=“path_to_client_config_file” -inputpath=“path_to_input_file” -logpath="path_to_log_folder"
``` 

Parameter Descriptions:
- For server and output party: use -mode="http" to disable TLS; the default enables it (which requires setup of certificate).
- For server: use -n_client_mal=num_of_malicious_client to enable malicious clients; the default assumes all are honest.
- For client: use -mode=honest to run client without malicious behavior. Default setting assumes client could act maliciously.
   
 **Note:** Servers and the output party must start before clients.


## Citation
If you find this work useful, please cite the following paper:

Jianan Su, Laasya Bangalore, Harel Berger, Jason Yi, Sophia Castor, Muthuramakrishnan Venkitasubramaniam, and Micah Sherr. “SCIF: Privacy-Preserving Statistics Collection with Input Validation and Full Security.” In Privacy Enhancing Technologies Symposium (PETS), 2025.

```bibtex
@inproceedings{scif2025,
  author = {Su, Jianan and Bangalore, Laasya and Berger, Harel and Yi, Jason and Castor, Sophia and Venkitasubramaniam, Muthuramakrishnan and Sherr, Micah},
  title = {{SCIF: Privacy-Preserving Statistics Collection with Input Validation and Full Security}},
  booktitle = {Privacy Enhancing Technologies Symposium (PETS)},
  year = {2025},
  month = jul
}
```

A pre-print of our paper [is available](https://eprint.iacr.org/2024/1821).



  
