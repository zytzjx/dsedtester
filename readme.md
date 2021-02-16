# Erase Storage

```

sudo sed -i 's/databases 16/databases 81/g' /etc/redis/redis.conf
```

## storage 0

tasklist    all data task data ([]string)


### Erase Disk Redis key

* starttasktime
* endtasktime
* errorcode
* processing
* SupportList      Features Supported List  

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

#### base errorcode 36000, all error code +36000

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


# Some Command line

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

```

sudo hdparm -g /dev/sdd

/dev/sdd:
SG_IO: bad/missing sense data, sb[]:  70 00 05 00 00 00 00 0a 00 00 00 00 20 00 01 cf 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
 geometry      = 36472/255/63, sectors = 585937500, start = 0



sudo hdparm --dco-identify /dev/sde

/dev/sde:
DCO Checksum verified.
DCO Revision: 0x0002
The following features can be selectively disabled via DCO:
        Transfer modes:
                 mdma0 mdma1 mdma2
                 udma0 udma1 udma2 udma3 udma4 udma5 udma6
        Real max sectors: 23437770752
        ATA command/feature sets:
                 SMART self_test error_log security PUIS HPA 48_bit
                 streaming FUA selective_test
                 WRITE_UNC_EXT
        SATA command/feature sets:
                 NCQ NZ_buffer_offsets interface_power_management SSP




sudo smartctl -a /dev/sdd
smartctl 7.1 2019-12-30 r5022 [x86_64-linux-5.4.0-42-generic] (local build)
Copyright (C) 2002-19, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Vendor:               SEAGATE
Product:              ST3300657SS
Revision:             000B
Compliance:           SPC-3
User Capacity:        300,000,000,000 bytes [300 GB]
Logical block size:   512 bytes
Rotation Rate:        15000 rpm
Form Factor:          3.5 inches
Logical Unit id:      0x5000c50074140483
Serial number:        6SJ0WSW10000N133AALV
Device type:          disk
Transport protocol:   SAS (SPL-3)
Local Time is:        Tue Jan 19 13:19:54 2021 PST
SMART support is:     Available - device has SMART capability.
SMART support is:     Enabled
Temperature Warning:  Enabled

=== START OF READ SMART DATA SECTION ===
SMART Health Status: OK
Current Drive Temperature:     0 C
Drive Trip Temperature:        0 C

Elements in grown defect list: 0

Error Counter logging not supported

Device does not support Self Test logging
```