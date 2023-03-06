#!/bin/bash

for i in {1..8}
do
	echo "Copying code to $i server"
	sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no "$netid@sp23-cs425-220$i.cs.illinois.edu" "rm -rf mp1 && mkdir mp1"
	# sshpass -p $UIUC_PASSWORD rsync --delete --ignore-existing -av -e ssh --exclude='venv' --exclude='*_log.txt'  --exclude='*_log.csv' --exclude='bin' --exclude='logs'../mp1 $netid@sp23-cs425-220$i.cs.illinois.edu:~/
	sshpass -p $UIUC_PASSWORD scp -o StrictHostKeyChecking=no -r cmd 3_config_prod.txt 8_config_prod.txt go.mod go.sum Makefile gentx.py "$netid@sp23-cs425-220$i.cs.illinois.edu:/home/$netid/mp1"
	sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no "$netid@sp23-cs425-220$i.cs.illinois.edu" "ls mp1"
done


for i in {1..8}
do
	echo "Running node$i"
	sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no -n -f "$netid@sp23-cs425-220$i.cs.illinois.edu" "cd mp1 && make && nohup python3 -u gentx.py 5 | ./bin/mp1_node node$i 8_config_prod.txt"&
done
