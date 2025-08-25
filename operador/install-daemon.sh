#!/bin/bash
set -e

DAEMON_NAME="operator"
GO_BINARY="/usr/local/bin/${DAEMON_NAME}"
SERVICE_FILE="/etc/systemd/system/${DAEMON_NAME}.service"
PROJECT_DIR="/home/ral/projetos/labOperators"
USER_NAME="ral"
GROUP_NAME="ral"

echo "=== Instalando ${DAEMON_NAME} daemon"

if [[ $EUID -ne 0 ]]; then
    echo "script deve ser executado como root"
    exit 1
fi

if [[ ! -f "./${DAEMON_NAME}" ]]; then
    echo "binario Go './${DAEMON_NAME}' não encontrado"
    exit 1
fi

install -o root -g root -m 755 "./${DAEMON_NAME}" "${GO_BINARY}"
echo "Instalado: ${GO_BINARY}"

# systemd service
# 
cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=${DAEMON_NAME}
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=0

[Service]
Type=simple
ExecStart=${GO_BINARY}
Restart=always
RestartSec=10
User=${USER_NAME}
Group=${GROUP_NAME}

Environment=GO_ENV=production
Environment=LOG_LEVEL=info
Environment=HOME=/home/${USER_NAME}

WorkingDirectory=${PROJECT_DIR}

StandardOutput=journal
StandardError=journal
SyslogIdentifier=${DAEMON_NAME}

LimitNOFILE=65536
LimitNPROC=32768

TimeoutStartSec=30
TimeoutStopSec=15
KillMode=mixed
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
EOF

echo "Service: ${SERVICE_FILE}"


mkdir -p "${PROJECT_DIR}"
chown "${USER_NAME}:${GROUP_NAME}" "${PROJECT_DIR}"
chmod 755 "${PROJECT_DIR}"
echo "diretorio pronto: ${PROJECT_DIR}"

systemctl daemon-reload
systemctl enable "${DAEMON_NAME}"

echo "servico habilitado para iniciar no boot"

cat <<EOF
Comandos:
  sudo systemctl start ${DAEMON_NAME}
  sudo systemctl stop ${DAEMON_NAME}

  logs:
  sudo systemctl status ${DAEMON_NAME}
  sudo journalctl -u ${DAEMON_NAME} -f
  sudo journalctl -u ${DAEMON_NAME} --no-pager -l

Testar manualmente o código:
  ${GO_BINARY}

atualizar o binario:
  o build -o ${DAEMON_NAME} main.go
  sudo systemctl stop ${DAEMON_NAME}
  sudo install -o root -g root -m 755 ${DAEMON_NAME} ${GO_BINARY}
  sudo systemctl start ${DAEMON_NAME}
EOF
