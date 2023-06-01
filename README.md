# SMC 
+ To test client end, go to the folder SMC/cmd/client and run command 

	go run main.go -cid="c1"
+ To test server end, go to the folder SMC/cmd/server and run command 
  
  go run main.go -port=":8080" -sid="s1"

## Status
+ Client end could splite single secret and send to servers. Secret, servers' urls and parameters for packed secret sharing can be read from config file.
+ Server end could receive share from different client. 

## To Do
+ Test aggregating shares and reconstruct sum of secrets.
+ Let server listen to client, aggregate shares and send to output party concurrently.
+ Change go.mod to import packed secret sharing from github 
  