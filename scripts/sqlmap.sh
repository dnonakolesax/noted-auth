#!/bin/bash

if [ ! -d "/opt/noted-stuff/sqlmap" ]; then
        echo "sqlmap not found, installing..."
        mkdir -p /opt/noted-stuff/sqlmap
        cd /opt/noted-stuff/sqlmap && git clone -b swagger-flag https://github.com/sqlmapproject/sqlmap.git
fi

python3 /opt/noted-stuff/sqlmap/sqlmap.py --swagger-file docs/swagger.yaml
