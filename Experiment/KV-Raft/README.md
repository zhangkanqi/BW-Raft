Implement Basic Raft With Key-value Store.
#### 1. 运行sever.go
cd ~/Go/BW-Raft/Experiment/KV-Raft/serve
```shell script
go run server.go -address 192.168.8.3:5000 -members 192.168.8.3:5000,192.168.8.6:5000
go run server.go -address 192.168.8.6:5000 -members 192.168.8.3:5000,192.168.8.6:5000
```
#### 2. 运行client.go
cd ~/Go/BW-Raft/Experiment/KV-Raft/serve
```shell script
go run client.go -cluster 192.168.8.3:5000,192.168.8.6:5000
```