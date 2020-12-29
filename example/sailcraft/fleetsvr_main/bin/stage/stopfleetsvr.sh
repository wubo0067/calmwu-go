if [ ! -f "pid" ]; then
        echo "pid not exist"
else
        fleetpid=$(cat pid)
        echo "PID: $fleetpid"

        if pinfo=$(ps -p $fleetpid); then
                if [[ $pinfo =~ "fleet" ]]; then
                        echo $pinfo
                        echo "$fleetpid is running, now killing it..."
                        kill -9 $fleetpid
                        rm -rf pid
                else
                        echo "$fleetpid is not fleetsvr"
                fi
        else
                echo "$fleetpid is not running"
        fi
fi