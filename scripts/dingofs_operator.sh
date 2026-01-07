#!/bin/bash

# 生成随机文件系统名称函数
generate_random_fsname() {
    local prefix="dingofs"
    local random_str=$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 8 | head -n 1)
    echo "${prefix}-${random_str}"
}

# 显示中文提示信息
echo "支持的操作:"
echo "创建文件系统|删除文件系统|设置文件系统配额|删除文件系统配额|设置目录配额|删除目录配额|挂载|热升级"
echo ""
echo  -n "DingoFS 输入: "

read user_input

# 判断指令
case "$user_input" in
    "创建文件系统")
        echo "执行创建文件系统操作..."
        FSNAME=$(generate_random_fsname)
        echo "分配名称: $FSNAME"
        if dingo create fs --fsname "$FSNAME"; then
            echo "文件系统 $FSNAME 创建成功！"
        else
            echo "创建失败！"
            exit 1
        fi
        ;;
    "删除文件系统")
        echo -n "请输入要删除的文件系统名称: "
        read FSNAME
        echo "执行删除文件系统操作..."
        if dingo delete fs --fsname "$FSNAME" --noconfirm; then
            echo "文件系统 $FSNAME 删除成功！"
        else
            echo "删除失败！"
            exit 1
        fi
        ;;
    "设置文件系统配额")
        echo -n "请输入文件系统名称: "
        read FSNAME
        echo -n "请输入配额大小 (G): "
        read QUOTA
        echo -n "请输入文件数: "
        read NUMS
        echo "执行设置文件系统配额操作..."
        if dingo config fs --fsname "$FSNAME" --capacity "$QUOTA" --inodes "$NUMS"; then
            echo "文件系统 $FSNAME 配额设置为 [${QUOTA}G, $NUMS] 成功！"
        else
            echo "设置配额失败！"
            exit 1
        fi
        ;;
    "删除文件系统配额")
        echo -n "请输入文件系统名称: "
        read FSNAME
        echo "执行删除文件系统配额操作..."
        if dingo config fs --fsname "$FSNAME" --capacity 0 --inodes 0; then
            echo "文件系统配额成功！"
        else
            echo "删除配额失败！"
            exit 1
        fi
        ;;
    "设置目录配额")
        echo -n "请输入文件系统名称: "
        read FSNAME
        echo -n "请输入目录路径 (例如 /data): "
        read DIRPATH
        echo -n "请输入配额大小 (G): "
        read QUOTA
        echo -n "请输入文件数: "
        read NUMS
        echo "执行设置目录配额操作..."
        if dingo quota set --fsname "$FSNAME" --path "$DIRPATH" --capacity "$QUOTA" --inodes "$NUMS"; then
            echo "目录 $DIRPATH 在文件系统 $FSNAME 的配额设置为 [${QUOTA}(G), $NUMS] 成功！"
        else
            echo "设置目录配额失败！"
            exit 1
        fi
        ;;
    "删除目录配额")
        echo -n "请输入文件系统名称: "
        read FSNAME
        echo -n "请输入目录路径 (例如 /data): "
        read DIRPATH
        echo "执行删除目录配额操作..."
        if dingo quota delete --fsname "$FSNAME" --path "$DIRPATH"; then
            echo "目录 $DIRPATH 在文件系统 $FSNAME 的配额删除成功！"
        else
            echo "删除目录配额失败！"
            exit 1
        fi
        ;;
    "挂载"|"热升级")
        echo -n "请输入dingo-fuse路径 (例如 /usr/local/bin/dingo-fuse): "
        read DINGOFUSE_PATH
        echo -n "请输入dingo-fuse配置文件路径 (例如 /opt/client.conf): "
        read DINGOFUSE_CONF
        echo -n "请输入文件系统名称: "
        read FSNAME
        echo -n "请输入挂载点 (例如 /mnt/dingofs): "
        read MOUNTPOINT
        echo "执行挂载操作..."
        nohup sudo ${DINGOFUSE_PATH} -f -o default_permissions -o allow_other -o fsname=${FSNAME} -o fstype=vfs_v2 -o user=curvefs -o conf=${DINGOFUSE_CONF} ${MOUNTPOINT} > /dev/null 2>&1 &
        sleep 10
        if ps -ef | grep -v grep | grep -q "$MOUNTPOINT"; then
            echo "文件系统 $FSNAME 成功挂载到 $MOUNTPOINT"
        else
            echo "挂载失败！"
            exit 1
        fi
        ;;
    *)
        echo "未知指令: $user_input"
        exit 1
        ;;
esac