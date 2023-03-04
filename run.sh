#!/bin/bash
# trap "echo BOOOO" SIGINT
# for i in {1..8}
# do
# 	echo "Running node$i"
# 	# sshpass -p $UIUC_PASSWORD scp -o StrictHostKeyChecking=no -r ./mp0 "$netid@sp23-cs425-220$i.cs.illinois.edu:/home/$netid"
# 	# sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no "$netid@sp23-cs425-220$i.cs.illinois.edu" "ls mp0"
#     python3 -u gentx.py 1 | ./bin/mp1_node node$i 8_config.txt > /dev/null 2>&1 &
# done

#!/bin/bash

for i in {1..3}
do
	echo "Copying code to $i server"
	sshpass -p $UIUC_PASSWORD rsync --delete --ignore-existing -av -e ssh --exclude='venv' --exclude='*_log.txt'  --exclude='*_log.csv' --exclude='bin' --exclude='logs'../mp1 $netid@sp23-cs425-220$i.cs.illinois.edu:~/
	sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no "$netid@sp23-cs425-220$i.cs.illinois.edu" "ls mp1"
done


for i in {1..3}
do
	echo "Running node$i"
	sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no "$netid@sp23-cs425-220$i.cs.illinois.edu" "sh -c 'cd mp1 && make && nohup python3 -u gentx.py 1 | ./bin/mp1_node node$i 3_config_prod.txt &' &"&
done


# for i in {1..8}
# do
# 	echo "SSHing to $i server"
# 	sshpass -p $UIUC_PASSWORD ssh -o StrictHostKeyChecking=no -n -f "$netid@sp23-cs425-220$i.cs.illinois.edu" "sh -c 'cd mp0 && make && nohup python3 -u generator.py 5 | ./bin/node node$i sp23-cs425-2209.cs.illinois.edu 4321 > /dev/null 2>&1 &'"
# done
