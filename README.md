# Cosmonitor
Cosmonitor is a comprehensive monitoring infrastructure for Cosmos-SDK nodes and validators, offering alerting capabilities and a Grafana dashboard for a scalable and robust monitoring experience.

## Requirements

* 1 or more nodes/validators: The Cosmos-SDK nodes or validators that you want to monitor
* 1 dedicated monitoring vps: A dedicated monitoring Virtual Private Server (VPS) meeting these minimum requirements:
    * CPU: 1 core
    * RAM: 1 GB
    * HDD/SSD: 20 GB
    * Ubuntu 22 (suggested)

## Architecture

![Architecture Schema](./architecture.drawio.png)

## Monitor

The monitor server monitors all the targeted nodes/validators Prometheus servers, sending alerting notifications directly to the chosen provider (ex. Telegram, email, etc.) in case of issues.

### Setup

1. Install Docker:

```
curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh && rm get-docker.sh
```

2. Get the repo and navigate inside the `monitor` folder.
3. Create and configure `./prometheus/target.json` according to the `target.sample.json`.
4. Create and configure `./alertmanager/alertmanager.yml` according to the `alertmanager.sample.yml`. In order to add additional alerting providers, refer to [this guide](https://prometheus.io/docs/alerting/latest/configuration/#receiver-integration-settings) .
5. Enable the firewall:

```
 sudo ufw allow ssh
 sudo ufw allow 9090/tcp
 sudo ufw enable
```

1. Start the monitor with `./start.sh` to initiate monitoring.

## Node/Validator

### Setup

1. Install Docker:

```
curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh && rm get-docker.sh
```

2. Get the repo and navigate inside the `node` folder.
3. Add the following rules to the firewall:

```
 sudo ufw allow from 172.16.0.0/12 to any port 26657 proto tcp # docker internal usage
 sudo ufw allow from 172.16.0.0/12 to any port 26660 proto tcp # docker internal usage
 sudo ufw allow from <MONITOR_IP> to any port 39090 proto tcp # expose prometheus 
 sudo ufw reload
```

4. Edit the `$HOME/.<CHAIN_FOLDER>/config/config.toml` and set:

```
laddr = "tcp://0.0.0.0:26657" # [rpc] section
namespace = "cometbft" # if not already set
```

5. Restart the chain service with `sudo systemctl restart <CHAIN_SERVICE_NAME>`.
6. Start the server with `./start.sh`,
