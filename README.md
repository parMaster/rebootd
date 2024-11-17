# rebootd - A simple reboot daemon for Linux
rebootd is a simple reboot daemon for Linux. It supposed to be run as systemd service and will reboot the system if network is unreachable for a certain amount of time (failed to GET a certain urls - comma separated list, default: https://www.google.com,https://www.cloudflare.com,https://www.amazon.com).
It will try to restart `systemd-resolved` at the first failed attempt.
It will try to reach the url every 15 minutes and if it fails 5 times in a row, it will reboot the system.

## Use case
Remote servers that are not reachable and need to be rebooted. This is a simple solution to reboot the server if it's not reachable.

## Installation

1. Clone the repository
2. Run `make install` to install the service
3. Enjoy

## Options

```shell
✗ ./rebootd --help
Usage:
  rebootd [OPTIONS]

Application Options:
      --dbg             show debug info [$DEBUG]
  -v                    Show version and exit
  -a=                   Number of failed attempts allowed before reboot (default: 5)
      --address=        Address list to check - comma separated (default: https://www.google.com,https://www.cloudflare.com,https://www.amazon.com)
  -i, --interval=       Interval between checks (default: 15m) [$INTERVAL]
  -r, --retry-interval= Interval between checks after failed attempt (default: 5m) [$RETRY_INTERVAL]

Help Options:
  -h, --help            Show this help message
```

## Uninstall

Run `make uninstall` to remove the service.
