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
|999|user cancel task|
|100|sanitize not support|
|123|unknow failed, not start erasing|
|250|user input disk transcation|
|10|sanitize verify failed|
|11|not find linuxname|
|12|not find sgName|
|13|create log file failed|


# Some Command line:
```
# format
sudo ./openSeaChest_Erase --progress format --formatUnit current -d /dev/sg5 --confirm this-will-erase-data
# format progress
sudo ./openSeaChest_Erase --progress format  -d /dev/sg5

sudo ./openSeaChest_Erase -d /dev/sg4 --ataSecureErase normal --confirm this-will-erase-data
sudo ./openSeaChest_SMART -d /dev/sg3 --smartCheck  #only support SATA

sudo ./openSeaChest_GenericTests -d /dev/sg4 --butterflyTest --minutes 1
sudo ./openSeaChest_GenericTests -d /dev/sg4 --randomTest --minutes 2
sudo ./openSeaChest_GenericTests -d /dev/sg4 --bufferTest --minutes 3

sudo ./openSeaChest_Erase -d /dev/sg4 --ataSecureErase normal --confirm this-will-erase-data  #take long time
```