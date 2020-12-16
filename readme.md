### Erase Storage

```
sudo sed -i 's/databases 16/databases 81/g' /etc/redis/redis.conf
```

### storage 0
tasklist    all data task data ([]string)

### Erase Disk Redis key
* starttasktime   
* endtasktime    
* errorcode 
* processing   

```

                     This      All      All     This                              Single  
Pass No. of          Pass   Passes   Passes     Pass              Est.     MB/      MB/   
 No. Passes Byte Complete Complete  Elapsed  Consume    Start   Finish   Second   Second  
---- ------ ---- -------- -------- -------- -------- -------- -------- -------- --------  
   1      1 0xff   0.159%   0.159% 00:00:05 00:00:05 10:38:01 00003141    93.18    93.18  

```
     paser this data to DataBase,  regex=(\d*\.\d*)%.*?(\d*\.\d*)%.*?(\d*\.\d*)$

|name      |       index|
|:-------|---------:|
|speed|9|
|start|7|
|time|5|
|est|8|
|progress|4|
|optime|time/est|


# Error code List
### base errorcode 36000, all error code +36000
|error code|means|
|----------|-----|
|100|sanitize not support|
|250|user input disk transcation|
|10|sanitize verify failed|
|11|not find linuxname|
|12|not find sgName|
|13|create log file failed|

