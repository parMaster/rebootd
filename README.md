# rebootd - A simple reboot daemon for Linux
rebootd is a simple reboot daemon for Linux. It supposed to be run as systemd service and will reboot the system if network is inreachable for a certain amount of time.

## Installation

1. Clone the repository
2. Run `make install` to install the service
3. Enjoy

## Options

```shell
Usage:
  rebootd [OPTIONS]

Application Options:
      --dbg             show debug info [$DEBUG]
  -v                    Show version and exit
  -a                    Number of failed attempts allowed before reboot (default: 3)
      --address=        Address to check (default: https://www.google.com)
  -i, --interval=       Interval between checks (default: 15m) [$INTERVAL]
  -r, --retry-interval= Interval between checks after failed attempt (default: 5m) [$RETRY_INTERVAL]

Help Options:
  -h, --help            Show this help message
```

## Uninstall

Run `make uninstall` to remove the service.
