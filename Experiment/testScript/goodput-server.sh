#!/bin/bash
add="192.168.8.6:500"
add1="192.168.8.6:500"
add2="192.168.8.6:600"
add3="192.168.8.6:700"
n0=0
n1=10
n2=20
n3=30
address="-address"
conn=","
num=$1
gpwrite="goodput-write"
out=".txt"

## go run server.go -address 192.168.8.3:5002 -members 192.168.8.3:5001,192.168.8.3:5002,192.168.8.3:5003,192.168.8.6:5004,192.168.8.6:5005,192.168.8.3:5006 >> ./A6.txt &
## go run server.go -address 192.168.8.3:5003 -members 192.168.8.3:5001,192.168.8.3:5002,192.168.8.3:5003,192.168.8.6:5004,192.168.8.6:5005,192.168.8.3:5006 >> ./A6.txt &
## go run server.go -address 192.168.8.3:5004 -members 192.168.8.3:5001,192.168.8.3:5002,192.168.8.3:5003,192.168.8.6:5004,192.168.8.6:5005,192.168.8.3:5006 >> ./A6.txt &
## go run server.go -address 192.168.8.3:5005 -members 192.168.8.3:5001,192.168.8.3:5002,192.168.8.3:5003,192.168.8.6:5004,192.168.8.6:5005,192.168.8.3:5006 >> ./A6.txt &

## 0-29
function getCluster() {
	cluster=$add$n0
	for((j=1;j<$num;j++));do {
		if [ $j -ge $n1 ]; then
			if [ $j -ge $n2 ]; then ## >=20 700
			{
				add=$add3
				k=`expr $j - $n2` ## 2x-20 [0,9]
			}
			else ## >=10 600
			{
				add=$add2
				k=`expr $j - $n1` ## 1x0-10 [0,9]
			}
			fi
		else ## <10 500
		{
			add=$add1
			k=$j
		}
		fi
		cluster=$cluster$conn$add$k
	}
	done
	echo $cluster
}

for((i=0;i<$num;i++));do {
	if [ $i -ge $n1 ]; then
		if [ $i -ge $n2 ]; then ## >=20 700
		{
			add=$add3
			k=`expr $i - $n2` ## 2x-20 [0,9]
		}
		else ## >=10 600
		{
			add=$add2
			k=`expr $i - $n1` ## 1x0-10 [0,9]
		}
		fi
	else ## <10 500
	{
		add=$add1
		k=$i
	}
	fi

	me=$add$k
	## echo $me
	cluster=$(getCluster)
	## echo $cluster
	output=$gpwrite$num$out
	go run server.go -address $me -members $cluster  &
} &
done
