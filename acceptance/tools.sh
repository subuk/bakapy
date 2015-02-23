
wait_port_listened(){
    portN="$1"
    timeout="5"
    progressSymbol="."
    i=0
    while true; do
        let i+=1
        netstat -tlnp 2>/dev/null |awk '{print $4}' |grep -q :19875$ && break
        echo -n $progressSymbol
        if [ "$i" -gt "$timeout" ];then
            echo "Error: timeout"
            exit 1
        fi
        sleep 1
    done
}

kill_and_wait(){
    pid="$1"
    progressSymbol="."
    kill $pid
    while true; do
        test ! -d /proc/${pid} && break
        echo -n $progressSymbol
        sleep 0.1
    done
}