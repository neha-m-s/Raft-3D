# Raft3D-Distributed-3D-Printer-Management-System

**Raft3D** is a distributed system for managing 3D printers, filaments, and print jobs, powered by the **Raft Consensus Algorithm**. It ensures strong consistency across multiple nodes without relying on an external database.

---

## Key Highlights

- **Raft-based Consensus**: Ensures leader election and state replication.
- **Printer Management**: Register and manage 3D printers.
- **Filament Tracking**: Monitor and update filament inventory.
- **Print Job Scheduling**: Track print job statuses and progress.
- **No External Database**: Uses Raft FSM for durable in-memory state.
- **Fault Tolerant**: Seamlessly handles leader failures and re-elections.

## Getting Started

### Prerequisites

Go 1.18+

### Install Dependencies

```
go mod tidy
```

### Quick Start (Windows)

Run the following to spin up all 3 nodes instantly:
run_raft_cluster.bat

This will open 3 separate terminals for:
- Node 1 (Leader) on ports 8000/9000
- Node 2 on ports 8001/9001
- Node 3 on ports 8002/9002

---

### Interaction with Raft3D Cluster

**1. Post a Printer**
```
curl -X POST http://localhost:8000/printers \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"p1\",\"company\":\"Creality\",\"model\":\"Ender3\"}"
```

**2. Post a Filament**
```
curl -X POST http://localhost:8000/filaments \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"f1\",\"type\":\"PLA\",\"color\":\"Red\",\"total_weight_in_grams\":1000,\"remaining_weight_in_grams\":1000}"
```

**3. Post a Print Job**
```
curl -X POST http://localhost:8000/printjobs \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"j1\",\"printer_id\":\"p1\",\"filament_id\":\"f1\",\"filepath\":\"/prints/cube.gcode\",\"print_weight_in_grams\":200,\"status\":\"queued\"}"
```

**4. Update Print Job Status**
```
curl -X POST "http://localhost:8000/printjobs/j1/status?status=running"
```

**5. Get All Print Jobs**
```
curl -X GET "http://localhost:8000/printjobs"
```

**6. Get All Printers**
```
curl -X GET "http://localhost:8000/printers"
```

**7. Get All Filaments**
```
curl -X GET "http://localhost:8000/filaments"
```
---

### Replication from Followers

**Get All Printers**
```
curl http://localhost:8001/printers
curl http://localhost:8002/printers
```

**Get All Filaments**
```
curl http://localhost:8001/filaments
curl http://localhost:8002/filaments
```

**Get All Print Jobs**
```
curl http://localhost:8001/printjobs
curl http://localhost:8002/printjobs
```
---

### Leader Election
1) You can simulate leader failure and verify leader election works as expected:
2) Start all 3 nodes using your run_raft_cluster.bat.
3) Use curl or Postman to send a POST request to localhost:8000 (assumed leader).
4) Kill the process running the current leader (e.g., close the terminal at port 9000).
5) The remaining nodes will automatically elect a new leader.
6) Retry the same curl command (adjusting the port to the new leader).
7) Use GET requests on any node (8001, 8002, etc.) to confirm the data is consistent.
