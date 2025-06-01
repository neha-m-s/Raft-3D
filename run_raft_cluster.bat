@echo off

REM Node 1 will bootstrap the cluster
start "Raft Node 1" cmd /k raft3d.exe --id node1 --http 127.0.0.1:8000 --raft 127.0.0.1:9000 --peers 127.0.0.1:9001,127.0.0.1:9002 --bootstrap

REM Node 2
start "Raft Node 2" cmd /k raft3d.exe --id node2 --http 127.0.0.1:8001 --raft 127.0.0.1:9001 --peers 127.0.0.1:9000,127.0.0.1:9002

REM Node 3
start "Raft Node 3" cmd /k raft3d.exe --id node3 --http 127.0.0.1:8002 --raft 127.0.0.1:9002 --peers 127.0.0.1:9000,127.0.0.1:9001
