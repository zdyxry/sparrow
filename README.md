# sparrow

Use [Raft](https://github.com/hashicorp/raft) to complete leader election in the cluster.

## Feature
* Auto config VIP on leader node
* Clock synchronization of all nodes in the cluster

## Install

```shell script
git clone git@github.com:zdyxry/sparrow.git
cd sparrow
go mod download
go build
```

## Usage

Node1:
```shell script
[root@install1 12:55:56 tmp]$./sparrow serve --config ./sparrow.toml
Using config file: ./sparrow.toml
INFO[0000] Create config                                
INFO[0000] Initialize communication                     
INFO[0000] Create transport                             
INFO[0000] Create raft structures                       
INFO[0000] Create raft cluster configuration            
INFO[0000] Delete IP success                             ip=172.17.17.14 link=port-storage
INFO[0000] Started                                      
INFO[0001] Network Current Leader IP                     Leader Address="172.17.17.13:10000"
INFO[0001] NTP current Leader IP                         Leader Address="172.17.17.13:10000"
INFO[0001] Render follower config file: %s&{172.17.17.13 false [192.168.64.1]} 
```

Node2:
```shell script
[root@install2 12:55:57 tmp]$./sparrow serve --config ./sparrow.toml 
Using config file: ./sparrow.toml
INFO[0000] Create config                                
INFO[0000] Initialize communication                     
INFO[0000] Create transport                             
INFO[0000] Create raft structures                       
INFO[0000] Create raft cluster configuration            
INFO[0000] Delete IP success                             ip=172.17.17.14 link=port-storage
INFO[0000] Started                                      
INFO[0001] Network Current Leader IP                     Leader Address="172.17.17.13:10000"
INFO[0001] NTP current Leader IP                         Leader Address="172.17.17.13:10000"
INFO[0001] Render follower config file: %s&{172.17.17.13 false [192.168.64.1]} 
```

Node3:
```shell script
[root@install3 12:55:58 tmp]$./sparrow serve --config ./sparrow.toml 
Using config file: ./sparrow.toml
INFO[0000] Create config                                
INFO[0000] Initialize communication                     
INFO[0000] Create transport                             
INFO[0000] Create raft structures                       
INFO[0000] Create raft cluster configuration            
INFO[0000] Delete IP success                             ip=172.17.17.14 link=port-storage
INFO[0000] Started                                      
INFO[0009] NTP leading                                  
INFO[0009] Render leader config file: %s&{ false [192.168.64.1]} 
INFO[0009] Network leading                              
INFO[0009] Add IP success                                ip=172.17.17.14 link=port-storage
```

## Changelog
* 2020.01.03 add NetworkManager
* 2020.01.12 add Raft support and NTPManager
* 2020.01.17 refactor manager and send gratuitous arp after set vip

## Todo
* Leader election priority


## Inspired by
* https://github.com/darxkies/virtual-ip
* https://github.com/otoolep/hraftd
