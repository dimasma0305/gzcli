BIN=`which gzcli`
CSV="https://docs.google.com/spreadsheets/d/<id>/gviz/tq?tqx=out:csv"
SUDO ?=
WATCHER_PID_FILE=.gzctf/.watcher.pid
GITPULL_PID_FILE=.gzctf/.gitpull.pid
WATCHER_LOG_FILE=.gzctf/watcher.log
PUBLIC_URL="<public-url>"

sync:
	${SUDO} ${BIN} sync

sync-and-update-game:
	${SUDO} ${BIN} sync --update-game

setup:
	@echo "Setting up CTF environment with Traefik..."
	@echo "Traefik network will be created by docker-compose"
	@echo "You can now run 'make platform-up' to start the platform"

platform-up:
	@echo "Starting CTF platform with docker compose..."
	(cd .gzctf && ${SUDO} docker compose up -d)

platform-down:
	@echo "Stopping CTF platform with docker compose..."
	(cd .gzctf && ${SUDO} docker compose down)

gzcli-start:
	${SUDO} ${BIN} script start

gzcli-stop:
	${SUDO} ${BIN} script stop

watch:
	${SUDO} ${BIN} watch start

watch-stop:
	${SUDO} ${BIN} watch stop

watch-status:
	${SUDO} ${BIN} watch status

watch-logs:
	${SUDO} ${BIN} watch logs

register-all-user:
	${SUDO} ${BIN} team create ${CSV}

register-all-user-and-send-email:
	${SUDO} ${BIN} team create ${CSV} --send-email

flush-cache:
	(cd .gzctf && ${SUDO} docker compose exec -uroot cache redis-cli FLUSHALL)

traefik-logs:
	(cd .gzctf && ${SUDO} docker compose logs -f traefik)

traefik-dashboard:
	@echo "Traefik dashboard should be available at: http://localhost:8080"
	@echo "Note: Make sure to configure dashboard access in production"

restart-traefik:
	(cd .gzctf && ${SUDO} docker compose restart traefik)

ssl-status:
	@echo "SSL certificates are automatically managed by Traefik"
	@echo "Check Traefik logs for SSL certificate status: make traefik-logs"

platform-logs:
	(cd .gzctf && ${SUDO} docker compose logs -f)

db-logs:
	(cd .gzctf && ${SUDO} docker compose logs -f db)

gzctf-logs:
	(cd .gzctf && ${SUDO} docker compose logs -f gzctf)

platform-clean:
	@echo "Stopping all services and cleaning up..."
	(cd .gzctf && ${SUDO} docker compose down -v) || true
	@echo "Note: Traefik network will be recreated automatically if needed"

# Convenient aliases
up: platform-up
down: platform-down
clean: platform-clean

help:
	@echo "GZCLI Makefile"
	@echo ""
	@echo "Setup & Platform Management:"
	@echo "  setup               - Initial setup (no manual network creation)"
	@echo "  platform-up         - Start platform with docker compose up -d"
	@echo "  platform-down       - Stop platform with docker compose down"
	@echo "  platform-clean      - Stop and remove all containers/volumes"
	@echo ""
	@echo "GZCli Management (alternative):"
	@echo "  gzcli-start         - Start platform using gzcli scripts"
	@echo "  gzcli-stop          - Stop platform using gzcli scripts"
	@echo ""
	@echo "Challenge Sync:"
	@echo "  sync                - Sync challenges"
	@echo "  sync-and-update-game - Sync challenges and update game"
	@echo ""
	@echo "File Watcher:"
	@echo "  watch               - Start file watcher daemon"
	@echo "  watch-stop          - Stop the watcher daemon"
	@echo "  watch-status        - Check watcher status"
	@echo "  watch-logs          - View watcher logs"
	@echo ""
	@echo "User Management:"
	@echo "  register-all-user   - Register users from CSV"
	@echo "  register-all-user-and-send-email - Create teams and send emails from CSV"
	@echo ""
	@echo "Traefik Operations:"
	@echo "  traefik-logs        - View Traefik logs"
	@echo "  traefik-dashboard   - Show Traefik dashboard info"
	@echo "  restart-traefik     - Restart Traefik service"
	@echo "  ssl-status          - Show SSL certificate status"
	@echo ""
	@echo "Logging:"
	@echo "  platform-logs       - View all platform service logs"
	@echo "  db-logs             - View database logs only"
	@echo "  gzctf-logs          - View GZCTF application logs only"
	@echo ""
	@echo "Utilities:"
	@echo "  flush-cache         - Clear Redis cache"
	@echo ""
	@echo "Convenient Aliases:"
	@echo "  up                  - Alias for platform-up"
	@echo "  down                - Alias for platform-down"
	@echo "  clean               - Alias for platform-clean"

# Default check interval in minutes
UPDATE_Check_MIN ?= 1440

list-updated:
	@find events -name "challenge.yml" -mmin -${UPDATE_Check_MIN} -printf "%TY-%Tm-%Td %TH:%TM:%TS %p\n" | sort -r
