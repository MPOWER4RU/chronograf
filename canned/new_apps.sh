#!/bin/sh

measurement=

# Usage info
show_help() {

cat << EOF
Usage: ${0##*/} MEASUREMENT
Generate new layout for MEASUREMENT.  File created will be named
MEASUREMENT_UUID.json with UUID being generated from the uuidgen command.

    -h          display this help and exit
EOF
}

while :; do
    case $1 in
        -h|-\?|--help)   # Call a "show_help" function to display a synopsis, then exit.
            show_help
            exit
            ;;
        *)               # Default case: If no more options then break out of the loop.
			measurement=$1
            break
    esac
    shift
done

if [ -z "$measurement" ]; then
    show_help
	exit
fi


UUID=$(uuidgen)
UUID="$(tr [A-Z] [a-z] <<< "$UUID")"
APP_FILE="$measurement"_"$UUID".json
echo Creating measurement file $APP_FILE
cat > $APP_FILE << EOF
 {
    "id": "$UUID",
 	"measurement": "$measurement",
 	"app": "User Facing Application Name",
 	"cells": [{
 		"x": 0,
 		"y": 0,
 		"w": 10,
 		"h": 10,
 		"queries": [{
 			"query": "select used_percent from disk",
 			"db": "telegraf",
 			"rp": "autogen"
		}]
 	}]
 }
EOF
