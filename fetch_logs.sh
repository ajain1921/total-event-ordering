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


rm -rf logs && mkdir logs
for i in {1..3}
do
	echo "Copying code from $i server"
	sshpass -p $UIUC_PASSWORD scp -o StrictHostKeyChecking=no -r "$netid@sp23-cs425-220$i.cs.illinois.edu:/home/$netid/mp1/node${i}_*" ./logs/
# shpass -p $UIUC_PASSWORD scp -o StrictHostKeyChecking=no -r "$netid@sp23-cs425-220$i.cs.illinois.edu:/home/$netid/mp1/node${i}_log.txt" " ./logs/

done

