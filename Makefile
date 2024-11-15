B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d)

# get current user name
USER=$(shell whoami)
# get current user group
GROUP=$(shell id -gn)

.DEFAULT_GOAL: build

build:
	go build -v --ldflags="-X main.version=$(REV)" 

install: build
	sudo systemctl stop rebootd || true
	sudo systemctl disable rebootd || true
	sudo rm -f /usr/local/bin/rebootd || true
	sudo cp ./rebootd /usr/local/bin/rebootd
	sudo cp ./rebootd.service /etc/systemd/system/rebootd.service
	sudo systemctl daemon-reload
	sudo systemctl enable rebootd
	sudo systemctl start rebootd

stop:
	sudo systemctl stop rebootd || true

uninstall: stop
	sudo systemctl disable rebootd || true
	sudo rm -f /usr/local/bin/rebootd || true
	sudo rm -f /etc/systemd/system/rebootd.service || true
	sudo systemctl daemon-reload