# dingo tool usage

A tool for DingoFS

- [dingo tool usage](#dingo-tool-usage)
  - [How to use dingo tool](#how-to-use-dingo-tool)
    - [Install](#install)
    - [Introduction](#introduction)
  - [Command](#command)
    - [version](#version)
    - [check](#check)
      - [check copyset](#check-copyset)
      - [check chunk](#check-chunk)
    - [create](#create)
      - [create fs](#create-fs)
      - [create topology](#create-topology)
    - [delete](#delete)
      - [delete fs](#delete-fs)
      - [delete metaserver](#delete-metaserver)
    - [list](#list)
      - [list copyset](#list-copyset)
      - [list fs](#list-fs)
      - [list mountpoint](#list-mountpoint)
      - [list partition](#list-partition)
      - [list topology](#list-topology)
      - [list dentry](#list-dentry)
    - [query](#query)
      - [query copyset](#query-copyset)
      - [query fs](#query-fs)
      - [query inode](#query-inode)
      - [query metaserver](#query-metaserver)
      - [query partition](#query-partition)
      - [query dirtree](#query-dirtree)
    - [status](#status)
      - [status mds](#status-mds)
      - [status metaserver](#status-metaserver)
      - [status etcd](#status-etcd)
      - [status copyset](#status-copyset)
      - [status cluster](#status-cluster)
    - [umount](#umount)
      - [umount fs](#umount-fs)
    - [usage](#usage)
      - [usage inode](#usage-inode)
      - [usage metadata](#usage-metadata)
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
      - [stats cluster](#stats-cluster)
      
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
wget https://github.com/dingodb/dingofs/blob/main/tools-v2/pkg/config/dingo.yaml
```
Please modify the `mdsAddr, mdsDummyAddr, etcdAddr` under `dingofs` in the template.yaml file as required

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
sage:  dingo status COMMAND [OPTIONS]

get the status of dingofs

Commands:
  cluster     get status of the dingofs
  copyset     status all copyset of the dingofs
  etcd        get the etcd status of dingofs
  mds         get status of mds
  metaserver  get metaserver status of dingofs

Global Flags:
      --conf string   config file (default is $HOME/.dingo/dingo.yaml or /etc/dingo/dingo.yaml)
      --help          print help
      --showerror     display all errors in command
      --verbose       show some extra info

Run 'dingo status COMMAND --help' for more information on a command.

Examples:
$ dingo status mds
```

In addition, this tool reads the configuration from `$HOME/.dingo/dingo.yaml` or `/etc/dingo/dingo.yaml` by default,
and can be specified by `--conf` or `-c`.

## Command

### version

show the version of dingo tool

Usage:

```shell
dingo --version
```

Output:

```shel
Version: 2.0.1
Build Date: 2025-03-03T11:20:11Z
Git Commit: 7982a3c7ef7bfe367f669621cad5814a26a3db92+CHANGES
Go Version: go1.21.11
OS / Arch: linux amd64
```

### check

#### check copyset

check copysets health in dingofs

Usage:

```shell
dingo check copyset --copysetid 1 --poolid 1
```

Output:

```shell
+------------+-----------+--------+--------+---------+
| COPYSETKEY | COPYSETID | POOLID | STATUS | EXPLAIN |
+------------+-----------+--------+--------+---------+
| 4294967297 |         1 |      1 | ok     |         |
+------------+-----------+--------+--------+---------+
```

#### check chunk

check chunk consistency under directory

Usage:

```shell
dingo check chunk --fsid 1 --threads 8
```

Output:

```shell
2025-06-04 21:13:09.937: check all file chunks under dirinode[1]
- fsid: [600] inodeId: [9486140] name: [no_dismount.txt] duplicate chunkid: [235060843] 
	chunkIndex:0	chunkId:235060843  compaction:0  offset:0  len:68  size:68  zero:false
	chunkIndex:0	chunkId:235060843  compaction:0  offset:0  len:68  size:68  zero:false
2025-06-04 21:13:14.035: check over, dirinode[1], directories[110], files[10002], chunks[20007], errorChunks[2]
```

### create

#### create fs

create a fs in dingofs

Usage:

```shell
dingo  create fs --fsname dingofs  --fstype s3 --s3.ak 1CzODWr3xuiIOTl80CGc --s3.sk NR3Tk3hLK6GjehsawFeLPzHRweqwdMAGVMQ8ik1S --s3.endpoint https://172.20.61.102:19000 --s3.bucketname yansp-bucket --s3.blocksize 4MiB --s3.chunksize 4MiB --mdsaddr 172.20.61.102:26700,172.20.61.103:26700,172.20.61.105:26700
```

Output:

```shell
+---------+---------+
| FSNAME  | RESULT  |
+---------+---------+
| dingofs | success |
+---------+---------+
```

#### create topology

create dingofs topology

Usage:

```shell
dingo create topology --clustermap topology.json
```

Output:

```shell
+-------------------+--------+-----------+--------+
|       NAME        |  TYPE  | OPERATION | PARENT |
+-------------------+--------+-----------+--------+
| pool2             | pool   | del       |        |
+-------------------+--------+           +--------+
| zone4             | zone   |           | pool2  |
+-------------------+--------+           +--------+
| **.***.***.**_3_0 | server |           | zone4  |
+-------------------+--------+-----------+--------+
```

### delete

#### delete fs

delete a fs from dingofs

Usage:

```shell
dingo delete fs --fsname dingofs
WARNING:Are you sure to delete fs dingofs?
please input [dingofs] to confirm: dingofs
```

Output:

```shell
+--------+-------------------------------------+
| FSNAME |               RESULT                |
+--------+-------------------------------------+
| dingofs| delete fs failed!, error is FS_BUSY |
+--------+-------------------------------------+
```

#### delete metaserver

delete metaserver from topology

Usage:

```shell
dingo delete metaserver --metaserverid 1
WARNING:Are you sure to delete metaserver 1?
please input [yes] to confirm: yes
```

Output:

```shell
+---------+
| RESULT  |
+---------+
| success |
+---------+
```

### list

#### list copyset

list all copyset info of the dingofs

Usage:

```shell
dingo list copyset
```

Output:

```shell
+------------+-----------+--------+-------+--------------------------------+------------+
|    KEY     | COPYSETID | POOLID | EPOCH |           LEADERPEER           | PEERNUMBER |
+------------+-----------+--------+-------+--------------------------------+------------+
| 4294967302 | 6         | 1      | 2     | id:1                           | 3          |
|            |           |        |       | address:"**.***.***.**:6801:0" |            |
+------------+-----------+        +-------+                                +------------+
| 4294967303 | 7         |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+                                +------------+
| 4294967304 | 8         |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+                                +------------+
| 4294967307 | 11        |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+--------------------------------+------------+
| 4294967297 | 1         |        | 1     | id:2                           | 3          |
|            |           |        |       | address:"**.***.***.**:6802:0" |            |
+------------+-----------+        +-------+                                +------------+
| 4294967301 | 5         |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+                                +------------+
| 4294967308 | 12        |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+--------------------------------+------------+
| 4294967298 | 2         |        | 1     | id:3                           | 3          |
|            |           |        |       | address:"**.***.***.**:6800:0" |            |
+------------+-----------+        +-------+                                +------------+
| 4294967299 | 3         |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+                                +------------+
| 4294967300 | 4         |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+                                +------------+
| 4294967305 | 9         |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+        +-------+                                +------------+
| 4294967306 | 10        |        | 1     |                                | 3          |
|            |           |        |       |                                |            |
+------------+-----------+--------+-------+--------------------------------+------------+
```

#### list fs

list all fs info in the dingofs

Usage:

```shell
dingo list fs
```

Output:

```shell
+----+-------+--------+-----------+---------+----------+-----------+----------+
| ID | NAME  | STATUS | BLOCKSIZE | FSTYPE  | SUMINDIR |   OWNER   | MOUNTNUM |
+----+-------+--------+-----------+---------+----------+-----------+----------+
| 2  | test1 | INITED | 1048576   | TYPE_S3 | false    | anonymous | 1        |
+----+-------+--------+-----------+         +----------+           +----------+
| 3  | test3 | INITED | 1048576   |         | false    |           | 0        |
+----+-------+--------+-----------+---------+----------+-----------+----------+
```

#### list mountpoint

list all mountpoint of the dingofs

Usage:

```shell
dingo list mountpoint
```

Output:

```shell
+------+-----------+-----------------------------------+
| FSID |  FSNAME   |            MOUNTPOINT             |
+------+-----------+-----------------------------------+
| 16   | dingofs02 | dingofs-103:0:/mnt/dingofs        |
+------+-----------+-----------------------------------+
```

#### list partition

list partition in dingofs by fsid

Usage:

```shell
dingo list partition
```

Output:

```shell
+-------------+------+--------+-----------+----------+----------+-----------+
| PARTITIONID | FSID | POOLID | COPYSETID |  START   |   END    |  STATUS   |
+-------------+------+--------+-----------+----------+----------+-----------+
| 14          | 2    | 1      | 10        | 1048676  | 2097351  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 20          |      |        |           | 7340732  | 8389407  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 13          |      |        | 11        | 0        | 1048675  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 16          |      |        |           | 3146028  | 4194703  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 22          |      |        |           | 9438084  | 10486759 | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 21          |      |        | 5         | 8389408  | 9438083  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 23          |      |        | 7         | 10486760 | 11535435 | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 24          |      |        |           | 11535436 | 12584111 | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 15          |      |        | 8         | 2097352  | 3146027  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 18          |      |        |           | 5243380  | 6292055  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 17          |      |        | 9         | 4194704  | 5243379  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 19          |      |        |           | 6292056  | 7340731  | READWRITE |
+-------------+------+        +-----------+----------+----------+-----------+
| 26          | 3    |        | 2         | 1048676  | 2097351  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 30          |      |        |           | 5243380  | 6292055  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 34          |      |        | 3         | 9438084  | 10486759 | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 29          |      |        | 4         | 4194704  | 5243379  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 32          |      |        |           | 7340732  | 8389407  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 35          |      |        | 5         | 10486760 | 11535435 | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 27          |      |        |           | 2097352  | 3146027  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 33          |      |        |           | 8389408  | 9438083  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 25          |      |        | 6         | 0        | 1048675  | READWRITE |
+-------------+      +        +           +----------+----------+-----------+
| 36          |      |        |           | 11535436 | 12584111 | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 28          |      |        | 8         | 3146028  | 4194703  | READWRITE |
+-------------+      +        +-----------+----------+----------+-----------+
| 31          |      |        | 9         | 6292056  | 7340731  | READWRITE |
+-------------+------+--------+-----------+----------+----------+-----------+
```

#### list topology

list the topology of the dingofs

Usage:

```shell
dingo list topology
```

Output:

```shell
+-------+-------+----------+-----------------+--------------+---------------------+
| POOL  | ZONE  | SERVERID |     SERVER      | METASERVERID |     METASERVER      |
+-------+-------+----------+-----------------+--------------+---------------------+
| pool1 | zone1 | 1        | dingofs-102_0_0 | 1            | 172.20.61.102:16800 |
+       +-------+----------+-----------------+--------------+---------------------+
|       | zone2 | 2        | dingofs-103_1_0 | 3            | 172.20.61.103:16800 |
+       +-------+----------+-----------------+--------------+---------------------+
|       | zone3 | 3        | dingofs-105_2_0 | 2            | 172.20.61.105:16800 |
+-------+-------+----------+-----------------+--------------+---------------------+
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

### query

#### query copyset

query copysets in dingofs

Usage:

```shell
dingo query copyset --copysetid 1 --poolid 1
```

Output:

```shell
+------------+-----------+--------+--------------------------------------+-------+
| copysetKey | copysetId | poolId |              leaderPeer              | epoch |
+------------+-----------+--------+--------------------------------------+-------+
| 4294967297 |     1     |   1    | id:2  address:"**.***.***.**:6802:0" |   1   |
+------------+-----------+--------+--------------------------------------+-------+
```

#### query fs

query fs in dingofs by fsname or fsid

Usage:

```shell
dingo query fs --fsname dingofs02
```

Output:

```shell
+----+-----------+--------+-----------+---------+----------+-----------+----------+
| ID |   NAME    | STATUS | BLOCKSIZE | FSTYPE  | SUMINDIR |   OWNER   | MOUNTNUM |
+----+-----------+--------+-----------+---------+----------+-----------+----------+
| 16 | dingofs02 | INITED | 1048576   | TYPE_S3 | false    | anonymous | 1        |
+----+-----------+--------+-----------+---------+----------+-----------+----------+
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

#### query metaserver

query metaserver in dingofs by metaserverid or metaserveraddr

Usage:

```shell
dingo query metaserver --metaserverid=1,2,3
```

Output:

```shell
+----+---------------------------------+---------------------+---------------------+-------------+
| ID |            HOSTNAME             |    INTERNALADDR     |    EXTERNALADDR     | ONLINESTATE |
+----+---------------------------------+---------------------+---------------------+-------------+
| 1  | dingofs-metaserver-593795a6376a | 172.20.61.102:16800 | 172.20.61.102:16800 | ONLINE      |
+----+---------------------------------+---------------------+---------------------+-------------+
| 2  | dingofs-metaserver-faae2af68f0b | 172.20.61.105:16800 | 172.20.61.105:16800 | ONLINE      |
+----+---------------------------------+---------------------+---------------------+-------------+
| 3  | dingofs-metaserver-1c34a633c971 | 172.20.61.103:16800 | 172.20.61.103:16800 | ONLINE      |
+----+---------------------------------+---------------------+---------------------+-------------+
```

#### query partition

query the copyset of partition

Usage:

```shell
dingo query partition --partitionid 14
```

Output:

```shell
+----+--------+-----------+--------+-----------------------+
| ID | POOLID | COPYSETID | PEERID |       PEERADDR        |
+----+--------+-----------+--------+-----------------------+
| 14 | 1      | 2         | 1      | 172.20.61.102:16800:0 |
+    +        +           +--------+-----------------------+
|    |        |           | 2      | 172.20.61.105:16800:0 |
+    +        +           +--------+-----------------------+
|    |        |           | 3      | 172.20.61.103:16800:0 |
+----+--------+-----------+--------+-----------------------+
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

### status

#### status mds

get status of mds

Usage:

```shell
dingo status mds
```

Output:

```shell
+---------------------+---------------------+---------+----------+
|        ADDR         |      DUMMYADDR      | VERSION |  STATUS  |
+---------------------+---------------------+---------+----------+
| 172.20.61.102:26700 | 172.20.61.102:27700 | unknown | follower |
+---------------------+---------------------+         +          +
| 172.20.61.105:26700 | 172.20.61.105:27700 |         |          |
+---------------------+---------------------+         +----------+
| 172.20.61.103:26700 | 172.20.61.103:27700 |         | leader   |
+---------------------+---------------------+---------+----------+
```

#### status metaserver

get status of metaserver

Usage:

```shell
dingo status metaserver
```

Output:

```shell
+---------------------+---------------------+---------+--------+
|    EXTERNALADDR     |    INTERNALADDR     | VERSION | STATUS |
+---------------------+---------------------+---------+--------+
| 172.20.61.103:16800 | 172.20.61.103:16800 | unknown | online |
+---------------------+---------------------+         +        +
| 172.20.61.105:16800 | 172.20.61.105:16800 |         |        |
+---------------------+---------------------+         +        +
| 172.20.61.102:16800 | 172.20.61.102:16800 |         |        |
+---------------------+---------------------+---------+--------+
```

#### status etcd

get status of etcd

Usage:

```shell
dingo status etcd
```

Output:

```shell
+---------------------+---------+----------+
|        ADDR         | VERSION |  STATUS  |
+---------------------+---------+----------+
| 172.20.61.103:22379 | 3.4.10  | follower |
+---------------------+         +----------+
| 172.20.61.105:22379 |         | leader   |
+---------------------+---------+----------+
| 172.20.61.102:22379 | unknown | offline  |
+---------------------+---------+----------+
```

#### status copyset

get status of copyset

Usage:

```shell
dingo status copyset
```

Output:

```shell
+------------+-----------+--------+--------+--------+---------+
| COPYSETKEY | COPYSETID | POOLID | STATUS | LOGGAP | EXPLAIN |
+------------+-----------+--------+--------+--------+---------+
| 4294967297 | 1         | 1      | ok     | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967306 | 10        |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967307 | 11        |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967308 | 12        |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967298 | 2         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967299 | 3         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967300 | 4         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967301 | 5         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967302 | 6         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967303 | 7         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967304 | 8         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967305 | 9         |        |        | 0      |         |
+------------+-----------+--------+--------+--------+---------+
```

#### status cluster

get status of cluster

Usage:

```shell
dingo status cluster
```

Output:

```shell
etcd:
+---------------------+---------+----------+
|        ADDR         | VERSION |  STATUS  |
+---------------------+---------+----------+
| 172.20.61.103:22379 | 3.4.10  | follower |
+---------------------+         +----------+
| 172.20.61.105:22379 |         | leader   |
+---------------------+         +----------+
| 172.20.61.102:22379 |         | follower |
+---------------------+---------+----------+
mds:
+---------------------+---------------------+---------+----------+
|        ADDR         |      DUMMYADDR      | VERSION |  STATUS  |
+---------------------+---------------------+---------+----------+
| 172.20.61.102:26700 | 172.20.61.102:27700 | unknown | follower |
+---------------------+---------------------+         +          +
| 172.20.61.105:26700 | 172.20.61.105:27700 |         |          |
+---------------------+---------------------+         +----------+
| 172.20.61.103:26700 | 172.20.61.103:27700 |         | leader   |
+---------------------+---------------------+---------+----------+
meataserver:
+---------------------+---------------------+---------+--------+
|    EXTERNALADDR     |    INTERNALADDR     | VERSION | STATUS |
+---------------------+---------------------+---------+--------+
| 172.20.61.103:16800 | 172.20.61.103:16800 | unknown | online |
+---------------------+---------------------+         +        +
| 172.20.61.105:16800 | 172.20.61.105:16800 |         |        |
+---------------------+---------------------+         +        +
| 172.20.61.102:16800 | 172.20.61.102:16800 |         |        |
+---------------------+---------------------+---------+--------+
copyset:
+------------+-----------+--------+--------+--------+---------+
| COPYSETKEY | COPYSETID | POOLID | STATUS | LOGGAP | EXPLAIN |
+------------+-----------+--------+--------+--------+---------+
| 4294967297 | 1         | 1      | ok     | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967306 | 10        |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967307 | 11        |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967308 | 12        |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967298 | 2         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967299 | 3         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967300 | 4         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967301 | 5         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967302 | 6         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967303 | 7         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967304 | 8         |        |        | 0      |         |
+------------+-----------+        +        +--------+---------+
| 4294967305 | 9         |        |        | 0      |         |
+------------+-----------+--------+--------+--------+---------+

Cluster health is:  ok
```

### umount

#### umount fs

umount fs from the dingofs cluster

Usage:

```shell
dingo umount fs --fsname dingofs --mountpoint dingofs-103:9009:/mnt/dingofs
```
you can get mountpoint from "dingo list mountpoint" command.

Output:

```shell
+----------+--------------------------------+---------+
|  FSNAME  |           MOUNTPOINT           | RESULT  |
+----------+--------------------------------+---------+
| dingofs  | dingofs-103:9009:/mnt/dingofs  | success |
+----------+--------------------------------+---------+
```

NOTE: 
umount fs command does't really umount dingo-fuse client,it's only remove mountpoint from mds when dingo-fuse abnormal exit. Please use linux command "umount /mnt/dingofs or fusemount3 -u /mnt/dingofs" to umount dingo-fuse. 

### usage

#### usage inode

get the inode usage of dingofs

Usage:

```shell
dingo usage inode
```

Output:

```shell
+------+----------------+-----+
| fsId |     fsType     | num |
+------+----------------+-----+
|  2   |   inode_num    |  3  |
|  2   | type_directory |  1  |
|  2   |   type_file    |  0  |
|  2   |    type_s3     |  2  |
|  2   | type_sym_link  |  0  |
|  3   |   inode_num    |  1  |
|  3   | type_directory |  1  |
|  3   |   type_file    |  0  |
|  3   |    type_s3     |  0  |
|  3   | type_sym_link  |  0  |
+------+----------------+-----+
```

#### usage metadata

get the usage of metadata in dingofs

Usage:

```shell
dingo usage metadata
```

Output:

```shell
+---------------------+---------+---------+---------+
|   METASERVERADDR    |  TOTAL  |  USED   |  LEFT   |
+---------------------+---------+---------+---------+
| 172.20.61.103:16800 | 2.0 TiB | 359 GiB | 1.6 TiB |
+---------------------+---------+---------+---------+
| 172.20.61.105:16800 | 2.0 TiB | 48 GiB  | 2.0 TiB |
+---------------------+---------+---------+---------+
| 172.20.61.102:16800 | 2.0 TiB | 326 GiB | 1.7 TiB |
+---------------------+---------+---------+---------+
```

### warmup

#### warmup add

warmup a file(directory), or given a list file contains a list of files(directories) that you want to warmup.

Usage:

```shell
dingo warmup add /mnt/dingofs/warmup
dingo warmup add --filelist /mnt/dingofs/warmup.list
```

> `dingo warmup add /mnt/dingofs/warmup` will warmup a file(directory).
> /mnt/dingofs/warmup.list

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

set quota to a directory

Usage:

```shell
dingo quota set --fsid 1 --path /quotadir --capacity 10 --inodes 100000
```
#### quota get

get fs quota for dingofs

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

------usage------ ----------fuse--------- -metaserver -blockcache ---object--
 cpu   mem   wbuf| ops   lat   read write| ops   lat | read write| get   put 
 0.5% 1108M    0 |   0     0     0     0 |   0     0 |   0     0 |   0     0 
 0.6% 1109M    0 |  18  0.40     0     0 |   0     0 |   0     0 |   0     0 
 0.6% 1109M    0 |  18  0.16     0     0 |   0     0 |   0     0 |   0     0 
 0.7% 1109M    0 |  18  0.15     0     0 |   0     0 |   0     0 |   0     0 
 0.7% 1109M    0 |  18  0.14     0     0 |   0     0 |   0     0 |   0     0 
 0.7% 1109M    0 |  18  0.16     0     0 |   0     0 |   0     0 |   0     0 
 0.7% 1109M    0 |  18  0.16     0     0 |   0     0 |   0     0 |   0     0 
```

#### stats cluster

show real time performance statistics of dingofs cluster

Usage:

```shell
# Show by fsid
dingo stats cluster --fsid 1
			
# Show by fsname
dingo stats cluster --fsname dingofs

# Show 3 times
dingo stats cluster --fsid 1 --count 3

# Show every 4 seconds
dingo stats cluster --fsid 1 --interval 4s

```
Output:

```shell
dingo stats cluster --fsname dingofs

----------fuse--------- ---------object--------
 read  ops  write  ops | get   ops   put   ops 
   0     0     0     0 |   0     0     0     0 
   0     0  3689K 3689 |   0     0  4096K    1 
   0     0  3220K 3220 |   0     0  4096K    1 
   0     0  3752K 3752 |   0     0  4096K    1 
 133M 1067  3078K 3078 |   0     0  4096K    1 
 240M 1921  2498K 2498 |   0     0     0     0 
 207M 1663  1713K 1713 |   0     0  4096K    1 
 254M 2033  2538K 2538 |   0     0     0     0 
 254M 2032  2506K 2506 |   0     0  4096K    1 
 237M 1899  2342K 2342 |   0     0  4096K    1 
 195M 1563  1715K 1715 |   0     0     0     0 
 207M 1662  1987K 1987 |   0     0  4096K    1 
 251M 2010  2599K 2599 |   0     0     0     0 
 245M 1967  2586K 2586 |   0     0  4096K    1 
```

## Comparison of old and new commands

### dingo fs

| old                            | new                        |
| ------------------------------ | -------------------------- |
| dingofs_tool check-copyset     | dingo check copyset     |
| dingofs_tool create-fs         | dingo create fs         |
| dingofs_tool create-topology   | dingo create topology   |
| dingofs_tool delete-fs         | dingo delete fs         |
| dingofs_tool list-copyset      | dingo list copyset      |
| dingofs_tool list-fs           | dingo list fs           |
| dingofs_tool list-fs           | dingo list mountpoint   |
| dingofs_tool list-partition    | dingo list partition    |
| dingofs_tool query-copyset     | dingo query copyset     |
| dingofs_tool query-fs          | dingo query fs          |
| dingofs_tool query-inode       | dingo query inode       |
| dingofs_tool query-metaserver  | dingo query metaserver  |
| dingofs_tool query-partition   | dingo query partition   |
| dingofs_tool status-mds        | dingo status mds        |
| dingofs_tool status-metaserver | dingo status metaserver |
| dingofs_tool status-etcd       | dingo status etcd       |
| dingofs_tool status-copyset    | dingo status copyset    |
| dingofs_tool status-cluster    | dingo status cluster    |
| dingofs_tool umount-fs         | dingo umount fs         |
| dingofs_tool usage-inode       | dingo usage inode       |
| dingofs_tool usage-metadata    | dingo usage metadata    |