#!/bin/bash

cd ..
cd protocol-status-plugin
go run main.go >> ../logs/protocol-status.log 2>&1 &
cd ..
cd server-status-plugin
go run main.go >> ../logs/server-status.log 2>&1 &