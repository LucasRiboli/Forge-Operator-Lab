#!/bin/bash

# Script de desinstalação do daemon operator-keda-like

set -e

DAEMON_NAME="operator"
GO_BINARY="/usr/local/bin/${DAEMON_NAME}"
SERVICE_FILE="/etc/systemd/system/${DAEMON_NAME}.service"
PROJECT_DIR="/home/ral/projetos/labOperators"
USER_NAME="ral"
GROUP_NAME="ral"

echo "Desinstalando ${DAEMON_NAME} daemon"

# Verificar se é executado como root
if [ "$EUID" -ne 0 ]; then
    echo "este script deve ser executado como root"
    exit 1
fi

echo "para o serviço"
if systemctl is-active --quiet "${DAEMON_NAME}"; then
    systemctl stop "${DAEMON_NAME}"
    echo "Serviço parado"
else
    echo "Serviço já estava parado"
fi

echo "desabilita servico"
if systemctl is-enabled --quiet "${DAEMON_NAME}" 2>/dev/null; then
    systemctl disable "${DAEMON_NAME}"
    echo "servico desabilitado"
else
    echo "servivo já estava desabilitado"
fi

if [ -f "$SERVICE_FILE" ]; then
    rm -f "$SERVICE_FILE"
    echo "servico removido: $SERVICE_FILE"
else
    echo "servico não encontrado"
fi

systemctl daemon-reload
echo "Systemd recarregado"

if [ -f "$GO_BINARY" ]; then
    rm -f "$GO_BINARY"
    echo "binario removido: $GO_BINARY"
else
    echo "binario não encontrado"
fi

if [ -d "$WORK_DIR" ]; then
    if [ -z "$(ls -A $WORK_DIR)" ]; then
        rmdir "$WORK_DIR"
        echo "Dir de trabalho removido: $WORK_DIR"
    else
        echo "Dir de trabalho não está vazio, mantendo: $WORK_DIR"
    fi
else
    echo "Dir de trabalho não encontrado"
fi

echo "Desinstalacão finalizada"
echo "Para verificar se foi completamente removido:"
echo "  sudo systemctl status ${DAEMON_NAME}       # Deve retornar 'not found'"
echo "  sudo journalctl -u ${DAEMON_NAME}          # Logs antigos (se houver)"
echo "  ls -la $GO_BINARY                          # Deve retornar 'not found'"
echo "Se quiser reinstalar no futuro, execute:"
echo "  sudo ./install-daemon.sh"