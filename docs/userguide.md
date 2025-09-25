# dingo tool usage

A tool for DingoFS

- [dingo tool usage](#dingo-tool-usage)
  - [How to use dingo tool](#how-to-use-dingo-tool)
    - [Install](#install)
    - [Introduction](#introduction)
  - [Command](#command)
    - [version](#version)
    - [create](#create)
      - [create fs](#create-fs)
      - [create subpath](#create-subpath)
    - [delete](#delete)
      - [delete fs](#delete-fs)
      - [delete cachemember](#delete-cachemember)
    - [list](#list)
      - [list fs](#list-fs)
      - [list mountpoint](#list-mountpoint)
      - [list dentry](#list-dentry)
      - [list cachegroup](#list-cachegroup)
      - [list cachemember](#list-cachemember)
    - [query](#query)
      - [query fs](#query-fs)
      - [query inode](#query-inode)
      - [query dirtree](#query-dirtree)
    - [set](#set)
      - [set cachemember](#set-cachemember)
    - [status](#status)
      - [status mds](#status-mds)
    - [umount](#umount)
      - [umount fs](#umount-fs)
    - [usage](#usage)
      - [usage fs](#usage-fs)
    - [warmup](#warmup)
      - [warmup add](#warmup-add)
    - [config](#config)
      - [config fs](#config-fs)
      - [config get](#config-get)
      - [config check](#config-check)
    - [quota](#quota)
      - [quota set](#quota-set)
      - [quota get](#quota-get)
      - [quota list](#quota-list)
      - [quota delete](#quota-delete)
      - [quota check](#quota-check)
    - [stats](#stats)
      - [stats mountpoint](#stats-mountpoint)
    - [unlock](#unlock)
      - [unlock cachemember](#unlock-cachemember)
      
## How to use dingo tool

### Install

install dingo tool

For obtaining binary package, please refer to:
[dingo tool binary compilation guide](https://github.com/dingodb/dingofs/blob/main/INSTALL.md)

```bash
chmod +x dingo
mv dingo /usr/bin/dingo
```

set configure file

```bash
wget https://raw.githubusercontent.com/dingodb/dingofs-tools/main/pkg/config/dingo.yaml
```
Please modify the `mdsAddr` under `dingofs` in the template.yaml file as required

configure file priority
environment variables(CONF=/opt/dingo.yaml) > default (~/.dingo/dingo.yaml)
```bash
mv dingo.yaml ~/.dingo/dingo.yaml
or
export CONF=/opt/dingo.yaml
```

### Introduction

Here's how to use the tool

```bash
dingo COMMAND [options]
```

When you are not sure how to use a command, --help can give you an example of use:

```bash
dingo COMMAND --help
```

For example:

```bash
dingo status mds --help
Usage:  dingo status mds [flags]

get status of mds

Flags:
      --format string            output format (json|plain) (default "plain")
      --mdsaddr string           mds address, should be like 10.220.32.1:6700,10.220.32.2:6700,10.220.32.3:6700
      --rpcretrydelay duration   rpc retry delay (default 200ms)
      --rpcretrytimes int32      rpc retry times (default 5)
      --rpctimeout duration      rpc timeout (default 30s)

Global Flags:
      --conf string   config file (default is $HOME/.dingo/dingo.yaml or /etc/dingo/dingo.yaml)
      --help          print help
      --showerror     display all errors in command
      --verbose       show some extra info

Examples:
$ dingo status mds
```

In addition, this tool reads the configuration from `$HOME/.dingo/dingo.yaml` or `/etc/dingo/dingo.yaml` by default,
and can be specified by `--conf`.

### Config file example
```shell
Examples:

global:
  httpTimeout: 50000ms
  rpcTimeout: 50000ms
  rpcRetryTimes: 3
  rpcRetryDelay: 200ms
  showError: false

dingofs:
  mdsAddr: 172.20.61.102:26700,172.20.61.103:26700,172.20.61.105:26700
  storagetype: s3 # s3 or rados
  s3:
    ak: UAJ1WIVF3NM5XRIL0OU2
    sk: X9MxdCZslPmXADljX140iiN6r81aGgCnO61wEA3L 
    endpoint: http://10.220.68.19:80 
    bucketname: dingofs-bucket
    blocksize: 4 mib
    chunksize: 64 mib
  rados:
    username: client.dingofs-rgw
    key: AQANAExo/ihMLBAAPL8AXgqfxwdraw8uoWyJig==
    mon: 10.220.69.5:3300,10.220.69.6:3300,10.220.69.8:3300
    poolname: rados.dingofs.data
    blocksize: 4 mib
    chunksize: 64 mib
```

## Command

### version

show the version of dingo tool

Usage:

```shell
dingo version
```

Output:

```shel
Version: 5.0.0
Build Date: 2025-09-26T10:37:14Z
Git Commit: 600e11eaf2559bebc9c76aab6cbcd652dc1111f7
Go Version: go1.24.3
OS / Arch: linux amd64
```
### create

#### create fs

create fs in dingofs cluster

Usage:

```shell
# store in s3
$ dingo create fs --fsname dingofs --storagetype s3 --s3.ak 1CzODWr3xuiIOTl80CGc --s3.sk NR3Tk3hLK6GjehsawFeLPzHRweqwdMAGVMQ8ik1S --s3.endpoint http://localhost:9000 --s3.bucketname dingofs-bucket --mdsaddr 172.20.61.102:26700,172.20.61.103:26700,172.20.61.105:26700

# store in rados
$ dingo create fs --fsname dingofs --storagetype rados --rados.username admin --rados.key AQDg3Y2h --rados.mon 10.220.32.1:3300,10.220.32.2:3300,10.220.32.3:3300 --rados.poolname pool1 --rados.clustername ceph --mdsaddr 172.20.61.102:26700,172.20.61.103:26700,172.20.61.105:26700
```

Output:

```shell
+-------+---------+--------+-------------+--------------------------------------+---------+
| FSID  | FSNAME  | STATUS | STORAGETYPE |                 UUID                 | RESULT  |
+-------+---------+--------+-------------+--------------------------------------+---------+
| 10016 | dingofs | NORMAL | S3          | 2b2312fb-1931-4d37-abaa-69a762f84e27 | success |
+-------+---------+--------+-------------+--------------------------------------+---------+
```

#### create subpath

create sub directory in dingofs 

Usage:

```shell
 dingo create subpath --fsid 1 --path /path1
 dingo create subpath --fsname dingofs --path /path1/path2
```

Output:

```shell
+---------+
| RESULT  |
+---------+
| success |
+---------+
```

### delete

#### delete fs

delete fs from dingofs cluster

Usage:

```shell
dingo delete fs --fsname dingofs
WARNING:Are you sure to delete fs dingofs?
please input [dingofs] to confirm: dingofs
```

Output:

```shell
+----------+--------+
|  FSNAME | RESULT  |
+---------+---------+
| dingofs | success |
+---------+---------+
```
#### delete cachemember

delete cachegroup member

Usage:

```shell
 dingo delete cachemember --group test_cache --ip 10.225.10.170 --port 10001
```

Output:

```shell
+---------+
| RESULT  |
+---------+
| success |
+---------+
```

#### list fs

list all fs info in dingofs cluster

Usage:

```shell
dingo list fs
```

Output:

```shell
+-------+---------------------------------------+--------+-----------+-----------+--------+---------------+-------------------------------------+----------+--------------------------------------+
| FSID  |                FSNAME                 | STATUS | BLOCKSIZE | CHUNKSIZE | MDSNUM |  STORAGETYPE  |               STORAGE               | MOUNTNUM |                 UUID                 |
+-------+---------------------------------------+--------+-----------+-----------+--------+---------------+-------------------------------------+----------+--------------------------------------+
| 10000 | dingofs-stability-main16-3clients-211 | NORMAL | 4194304   | 67108864  | 2      | S3(HASH 1024) | http://10.220.88.21:8080/dingofs    | 0        | b674a17e-3899-44b5-850a-429fe93d9d9e |
+-------+---------------------------------------+--------+-----------+-----------+--------+               +-------------------------------------+----------+--------------------------------------+
| 10002 | dingofs-stability-main16-3clients-212 | NORMAL | 4194304   | 67108864  | 2      |               | http://10.220.88.21:8080/dingofs-sx | 0        | 7e93e886-cc60-470b-ac8c-ea15eaa42339 |
+-------+---------------------------------------+--------+-----------+-----------+--------+               +-------------------------------------+----------+--------------------------------------+
| 10004 | dingofs-stability-main16-3clients-213 | NORMAL | 4194304   | 67108864  | 2      |               | http://10.220.88.21:8080/dingofs-sx | 0        | a2514bc9-5b25-44c8-99b1-35fe67d80287 |
+-------+---------------------------------------+--------+-----------+-----------+--------+               +-------------------------------------+----------+--------------------------------------+
| 10006 | dingofs-stability-main16-3clients-224 | NORMAL | 4194304   | 67108864  | 2      |               | http://10.220.88.21:8080/dingofs-sx | 1        | 6afcc601-14c1-4d17-9c2a-5fcfde0260e0 |
+-------+---------------------------------------+--------+-----------+-----------+--------+               +-------------------------------------+----------+--------------------------------------+
| 10008 | dingofs-stability-main16-3clients-250 | NORMAL | 4194304   | 67108864  | 3      |               | http://10.220.88.21:8080/dingofs-sx | 2        | b8e512ed-7c2d-42d2-9ab6-4f3fede60b75 |
+-------+---------------------------------------+--------+-----------+-----------+--------+---------------+-------------------------------------+----------+--------------------------------------+
```

#### list mountpoint

list all mountpoint of the dingofs

Usage:

```shell
dingo list mountpoint
```

Output:

```shell
[yansp@dingofs-6 dingofs-tools]$ dingo list mountpoint  --mdsaddr 10.220.32.16:6900,10.220.32.17:6900,10.220.32.18:6900
+-------+---------------------------------------+--------------------------------------+------------------------------------+-------+
| FSID  |                FSNAME                 |               CLIENTID               |             MOUNTPOINT             |  CTO  |
+-------+---------------------------------------+--------------------------------------+------------------------------------+-------+
| 10006 | dingofs-stability-main16-3clients-224 | 8c8b443c-92bf-4ef3-9022-366e966fcb1d | ubuntu3:10020:/mnt/cluster/dingofs4| false |
+-------+---------------------------------------+--------------------------------------+------------------------------------+       +
| 10008 | dingofs-stability-main16-3clients-250 | 5404bbdc-bfca-4cf4-bd16-895a6dcccc46 | ubuntu1:10020:/mnt/cluster/dingofs4|       |
+       +                                       +--------------------------------------+------------------------------------+       +
|       |                                       | b1b9cbe3-d93c-47f7-9d38-bb9025176c54 | ubuntu2:10020:/mnt/cluster/dingofs4|       |
+-------+---------------------------------------+--------------------------------------+------------------------------------+-------+
```

#### list dentry

list directory dentry

Usage:

```shell
dingo list dentry --fsid 1 --inodeid 8393046
```

Output:

```shell
+------+----------+------+---------+----------------+
| FSID | INODEID  | NAME | PARENT  |      TYPE      |
+------+----------+------+---------+----------------+
| 2    | 9441683  | c48  | 8393046 | TYPE_S3        |
+------+----------+------+---------+----------------+
| 2    | 11538710 | c75  | 8393046 | TYPE_S3        |
+------+----------+------+---------+----------------+
| 2    | 11538673 | d46  | 8393046 | TYPE_DIRECTORY |
+------+----------+------+---------+----------------+
| 2    | 4336     | f43  | 8393046 | TYPE_S3        |
+------+----------+------+---------+----------------+
| 2    | 9441788  | ld4  | 8393046 | TYPE_SYM_LINK  |
+------+----------+------+---------+----------------+
```

#### list cachegroup

list all remote cache groups

Usage:

```shell
dingo list cachegroup
```

Output:

```shell
+---------+
|  GROUP  |
+---------+
| group_1 |
+---------+
```

#### list cachemember

list cachegroup members

Usage:

```shell
dingo list cachemember --group group_1
```

Output:

```shell
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
|               MEMBERID               |      IP      | PORT  | WEIGHT | LOCKED |       CREATE TIME       |    LAST ONLINE TIME     |  STATE  |               GROUP               |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 60b14376-7039-49fa-afee-47cbd318497c | 10.220.32.18 | 30020 | 100    | true   | 2025-09-22 11:36:48.000 | 2025-09-23 16:09:13.831 | offline |                                   |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 178c0bb7-7d9b-4ec2-9d54-f068db3b5e0e | 10.220.32.17 | 30020 | 100    | true   | 2025-09-22 11:36:47.000 | 2025-09-23 16:09:23.365 | offline |                                   |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 01a67188-08cd-4cf3-95fa-532b1ffd106e | 10.220.32.16 | 30020 | 100    | true   | 2025-09-22 11:36:47.000 | 2025-09-23 16:08:09.731 | offline |                                   |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| a1d3e3d9-c403-4b42-a46e-959b4f0b0403 | 10.220.32.16 | 30020 | 100    | true   | 2025-09-23 16:11:19.000 | 2025-09-26 12:25:31.141 | online  | dingofs-remote-cache-stability-v2 |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 39521621-78db-4365-8c68-d67827453ccc | 10.220.32.18 | 30020 | 100    | true   | 2025-09-23 16:11:19.000 | 2025-09-26 12:25:31.072 | online  | dingofs-remote-cache-stability-v2 |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 4a45a03a-7077-49d4-a58c-8e8da0f47006 | 10.220.32.17 | 30020 | 100    | true   | 2025-09-23 16:11:19.000 | 2025-09-26 12:25:31.321 | online  | dingofs-remote-cache-stability-v2 |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
```

### query

#### query fs

query fs in dingofs by fsname or fsid

Usage:

```shell
dingo query fs --fsid 10000
```

Output:

```shell
+-------+---------------------------------------+--------+-----------+-----------+--------+---------------+----------------------------------+----------+--------------------------------------+
| FSID  |                FSNAME                 | STATUS | BLOCKSIZE | CHUNKSIZE | MDSNUM |  STORAGETYPE  |             STORAGE              | MOUNTNUM |                 UUID                 |
+-------+---------------------------------------+--------+-----------+-----------+--------+---------------+----------------------------------+----------+--------------------------------------+
| 10000 | dingofs-stability-main16-3clients-211 | NORMAL | 4194304   | 67108864  | 2      | S3(HASH 1024) | http://10.220.88.21:8080/dingofs | 0        | b674a17e-3899-44b5-850a-429fe93d9d9e |
+-------+---------------------------------------+--------+-----------+-----------+--------+---------------+----------------------------------+----------+--------------------------------------+
```

#### query inode

query the inode of fs

Usage:

```shell
dingo query inode --fsid 2 --inodeid 5243380
```

Output:

```shell
+-------+----------+-----------+---------+-------+--------+
| FSID  | INODEID  |  LENGTH   |  TYPE   | NLINK | PARENT |
+-------+----------+-----------+---------+-------+--------+
|   2   | 5243380  | 352321536 | TYPE_S3 |   1   |  [1]   |
+-------+----------+-----------+---------+-------+--------+
```

#### query dirtree

recursive query parent inode

Usage:

```shell
dingo query dirtree --fsid 1 --inodeid 7344525
```

Output:

```shell
-- name  path:	/workunits/suites/tmp.roZYcz7Ln7/p9/dc/d1a/d1b/d21/d34
-- inode path:	1/8390147/10487316/9441617/2101392/7344488/6295896/6295898/7344505/7344525
```
### set

#### set cachemember

set remote cachegroup member attribute

Usage:

```shell
dingo set cachemember --memberid 3 --weight 40
```

Output:

```shell
+---------+
| RESULT  |
+---------+
| success |
+---------+
```

### status

#### status mds

get status of mds

Usage:

```shell
dingo status mds
```

Output:

```shell
+------+------------------+--------+-------------------------+-------------+
|  ID  |       ADDR       | STATE  |    LAST ONLINE TIME     | ONLINESTATE |
+------+------------------+--------+-------------------------+-------------+
| 1001 | 10.220.69.6:7400 | NORMAL | 2025-09-26 13:39:40.784 | online      |
+------+------------------+--------+-------------------------+-------------+
| 1002 | 10.220.69.6:7401 | NORMAL | 2025-09-26 13:39:44.747 | online      |
+------+------------------+--------+-------------------------+-------------+
| 1003 | 10.220.69.6:7402 | NORMAL | 2025-09-26 13:39:45.279 | online      |
+------+------------------+--------+-------------------------+-------------+
```

### umount

#### umount fs

umount fs from the dingofs cluster

Usage:

```shell
dingo umount fs --fsname dingofs --clientid 0cbe1e76-0afe-435b-9e60-1af57c836a3e
```
you can get mountpoint from "dingo list mountpoint" command.

Output:

```shell
+---------+--------------------------------------+---------+
| FSNAME  |               CLIENTID               | RESULT  |
+---------+--------------------------------------+---------+
| dingofs | 0cbe1e76-0afe-435b-9e60-1af57c836a3e | success |
+---------+--------------------------------------+---------+
```

NOTE: 
umount fs command does't really umount dingo-fuse client,it's only remove mountpoint from mds when dingo-fuse abnormal exit. Please use linux command "umount /mnt/dingofs or fusemount3 -u /mnt/dingofs" to umount dingo-fuse. 

### usage

#### usage fs

get the usage of fs in dingofs cluster

Usage:

```shell
dingo usage fs
```

Output:

```shell
+--------+---------------------------------------+-------------+-------+
|  FSID  |                FSNAME                 |    USED     | IUSED |
+--------+---------------------------------------+-------------+-------+
| 10000  | dingofs-stability-main16-3clients-211 | 44859       | 109   |
+--------+---------------------------------------+-------------+-------+
| 10002  | dingofs-stability-main16-3clients-212 | 0           | 1     |
+--------+---------------------------------------+-------------+-------+
| 10004  | dingofs-stability-main16-3clients-213 | 0           | 1     |
+--------+---------------------------------------+-------------+-------+
| 10006  | dingofs-stability-main16-3clients-224 | 30570228489 | 406   |
+--------+---------------------------------------+-------------+-------+
| 10008  | dingofs-stability-main16-3clients-250 | 20380152326 | 271   |
+--------+---------------------------------------+-------------+-------+
| TOTAL  |                   -                   | 50950425674 |  788  |
+--------+---------------------------------------+-------------+-------+

```

### warmup

#### warmup add

warmup a file(directory), or given a list file contains a list of files(directories) that you want to warmup.

Usage:

```shell
dingo warmup add /mnt/dingofs/warmup
dingo warmup add --filelist /mnt/dingofs/warmup.list
```

### config
#### config fs

config fs quota for dingofs

Usage:

```shell
dingo config fs --fsid 1 --capacity 100
dingo config fs --fsname dingofs --capacity 10 --inodes 1000000000
```
#### config get

get fs quota for dingofs

Usage:

```shell
dingo config get --fsid 1
dingo config get --fsname dingofs
```
Output:

```shell
+------+---------+----------+------+------+---------------+-------+-------+
| FSID | FSNAME  | CAPACITY | USED | USE% |    INODES     | IUSED | IUSE% |
+------+---------+----------+------+------+---------------+-------+-------+
| 2    | dingofs | 10 GiB   | 0 B  | 0    | 1,000,000,000 | 0     | 0     |
+------+---------+----------+------+------+---------------+-------+-------+
```

#### config check

check quota of fs

Usage:

```shell
dingo config check --fsid 1
dingo config check --fsname dingofs
```
Output:

```shell
+------+----------+-----------------+---------------+---------------+-----------+-------+-----------+---------+
| FSID |  FSNAME  |    CAPACITY     |     USED      |   REALUSED    |  INODES   | IUSED | REALIUSED | STATUS  |
+------+----------+-----------------+---------------+---------------+-----------+-------+-----------+---------+
| 1    | dingofs  | 107,374,182,400 | 1,083,981,835 | 1,083,981,835 | unlimited | 9     | 9         | success |
+------+----------+-----------------+---------------+---------------+-----------+-------+-----------+---------+
```

### quota
#### quota set

set quota to directory

Usage:

```shell
dingo quota set --fsid 1 --path /quotadir --capacity 10 --inodes 100000
```
#### quota get

get fs quota

Usage:

```shell
dingo quota get --fsid 1 --path /quotadir
dingo quota get --fsname dingofs --path /quotadir
```
Output:

```shell
+----------+------------+----------+------+------+------------+-------+-------+
|    ID    |    PATH    | CAPACITY | USED | USE% |   INODES   | IUSED | IUSE% |
+----------+------------+----------+------+------+------------+-------+-------+
| 10485760 | /quotadir1 | 10 GiB   | 6 B  | 0    | 20,000,000 | 1     | 0     |
+----------+------------+----------+------+------+------------+-------+-------+
```
#### quota list

list all directory quotas of fs

Usage:

```shell
dingo quota list --fsid 1
dingo quota list --fsname dingofs
```

Output:

```shell
+----------+------------+----------+------+------+------------+-------+-------+
|    ID    |    PATH    | CAPACITY | USED | USE% |   INODES   | IUSED | IUSE% |
+----------+------------+----------+------+------+------------+-------+-------+
| 10485760 | /quotadir1 | 10 GiB   | 6 B  | 0    | 20,000,000 | 1     | 0     |
+----------+------------+----------+------+------+------------+-------+-------+
| 2097152  | /quotadir2 | 100 GiB  | 0 B  | 0    | unlimited  | 0     |       |
+----------+------------+----------+------+------+------------+-------+-------+
```

#### quota delete

delete quota of a directory

Usage:

```shell
dingo quota delete --fsid 1 --path /quotadir
```
#### quota check

check quota of a directory

Usage:

```shell
dingo quota check --fsid 1 --path /quotadir
dingo quota check --fsid 1 --path /quotadir --repair
```


Output:

```shell
+----------+------------+----------------+------+----------+------------+-------+-----------+---------+
|    ID    |    NAME    |    CAPACITY    | USED | REALUSED |   INODES   | IUSED | REALIUSED | STATUS  |
+----------+------------+----------------+------+----------+------------+-------+-----------+---------+
| 10485760 | /quotadir | 10,737,418,240 | 22   | 22       | 20,000,000 | 2     | 22        | success |
+----------+------------+----------------+------+----------+------------+-------+-----------+---------+

or

+----------+------------+----------------+------+----------+------------+-------+-----------+--------+
|    ID    |    NAME    |    CAPACITY    | USED | REALUSED |   INODES   | IUSED | REALIUSED | STATUS |
+----------+------------+----------------+------+----------+------------+-------+-----------+--------+
| 10485760 | /quotadir | 10,737,418,240 | 22   | 33       | 20,000,000 | 2     | 3         | failed |
+----------+------------+----------------+------+----------+------------+-------+-----------+--------+
```

### stats
#### stats mountpoint

show real time performance statistics of dingofs mountpoint

Usage:

```shell
# normal
dingo stats mountpoint /mnt/dingofs
			
# fuse metrics
dingo stats mountpoint /mnt/dingofs --schema f

# s3 metrics
dingo stats mountpoint /mnt/dingofs --schema o

# More metrics
dingo stats mountpoint /mnt/dingofs --verbose

# Show 3 times
dingo stats mountpoint /mnt/dingofs --count 3

# Show every 4 seconds
dingo stats mountpoint /mnt/dingofs --interval 4s

```
Output:

```shell
dingo stats mountpoint /mnt/dingofs

------usage------ ----------fuse--------- ----blockcache--- ---object-- ------remotecache------
 cpu   mem   used| ops   lat   read write| load stage cache| get   put | load stage cache  hit 
 525% 4691M 2688K|   0     0     0     0 |   0     0     0 |   0     0 |   0     0     0   0.0%
 526% 4691M 1664K|1433  5.52   177M   95M|   0     0     0 |   0    96M| 453M    0    95M 99.4%
 527% 4691M 1152K|1418  5.71   157M   75M|   0     0     0 |   0    76M| 405M    0    76M 99.6%
 527% 4692M   64K|1531  5.24   189M   86M|   0     0     0 |   0    87M| 428M    0    86M 99.8%
 535% 4692M   64K|1415  5.55   180M   93M|   0     0     0 |   0    93M| 424M    0    93M 99.5%
 535% 4693M 1536K|1404  5.62   172M   96M|   0     0     0 |   0    95M| 396M    0    95M 99.5%
 537% 4692M 1152K|1420  5.55   171M   83M|   0     0     0 |   0    83M| 381M    0    84M 99.6%
 537% 4692M    0 |1303  5.92   170M   90M|   0     0     0 |   0    92M| 390M    0    90M 99.4%
 529% 4692M 2752K|1159  6.87   160M   81M|   0     0     0 |   0    79M| 391M    0    79M 99.5%
 528% 4692M 1600K|1372  5.87   166M   83M|   0     0     0 |   0    84M| 383M    0    86M 99.5%
 530% 4692M 3584K|1428  5.63   168M   79M|   0     0     0 |   0    77M| 435M    0    78M 99.4%
 528% 4692M    0 |1161  6.85   159M   71M|   0     0     0 |   0    74M| 363M    0    72M 99.3%
 500% 4692M    0 | 500  17.9    74M   37M|   0     0     0 |   0    37M| 167M    0    37M 99.6%
 490% 4692M 1664K|1113  7.35   146M   82M|   0     0     0 |   0    80M| 360M    0    80M 99.1%
 488% 4692M  640K|1431  5.53   167M   86M|   0     0     0 |   0    87M| 440M    0    87M 99.3%
 488% 4692M 1088K|1413  5.49   198M   92M|   0     0     0 |   0    92M| 441M    0    92M 99.6%
```

### unlock
#### unlock cachemember

Unbind the cachemember ID with the IP and Port

Usage:

```shell
$ dingo unlock cachemember  --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8 --ip 10.220.69.6 --port 10001
```
Output:

```shell
dingo  list cachemember
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
|               MEMBERID               |      IP      | PORT  | WEIGHT | LOCKED |       CREATE TIME       |    LAST ONLINE TIME     |  STATE  |               GROUP               |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 60b14376-7039-49fa-afee-47cbd318497c | 10.220.32.18 | 30020 | 100    | true   | 2025-09-22 11:36:48.000 | 2025-09-23 16:09:13.831 | offline |                                   |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+

dingo unlock cachemember --memberid 60b14376-7039-49fa-afee-47cbd318497c --ip 10.220.32.18 --port 30020
+---------+
| RESULT  |
+---------+
| success |
+---------+

dingo  list cachemember 
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
|               MEMBERID               |      IP      | PORT  | WEIGHT | LOCKED |       CREATE TIME       |    LAST ONLINE TIME     |  STATE  |               GROUP               |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
| 60b14376-7039-49fa-afee-47cbd318497c | 10.220.32.18 | 30020 | 100    | false  | 2025-09-22 11:36:48.000 | 2025-09-23 16:09:13.831 | offline |                                   |
+--------------------------------------+--------------+-------+--------+--------+-------------------------+-------------------------+---------+-----------------------------------+
```