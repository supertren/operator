#!/bin/bash
# =============================================================================
# Script 1: Setup WSL2 environment for hello-operator development
# Run this script INSIDE Ubuntu WSL2 after installation
# =============================================================================

set -e

echo "=================================================="
echo " Hello Operator - WSL2 Environment Setup"
echo "=================================================="

# --- Go 1.22 ---
echo ""
echo "[1/4] Installing Go 1.22..."
GO_VERSION="1.22.5"
wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz

if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
  echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
fi
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
go version
echo "Go instalado correctamente."

# --- operator-sdk ---
echo ""
echo "[2/4] Installing operator-sdk v1.34.1..."
ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; esac)
OS=$(uname | awk '{print tolower($0)}')
curl -sL "https://github.com/operator-framework/operator-sdk/releases/download/v1.34.1/operator-sdk_${OS}_${ARCH}" \
  -o /tmp/operator-sdk
chmod +x /tmp/operator-sdk
sudo mv /tmp/operator-sdk /usr/local/bin/operator-sdk
operator-sdk version
echo "operator-sdk instalado correctamente."

# --- oc CLI ---
echo ""
echo "[3/4] Installing oc CLI..."
wget -q "https://mirror.openshift.com/pub/openshift-v4/clients/ocp/stable/openshift-client-linux.tar.gz" \
  -O /tmp/oc.tar.gz
tar -xzf /tmp/oc.tar.gz -C /tmp oc kubectl
sudo mv /tmp/oc /usr/local/bin/oc
sudo mv /tmp/kubectl /usr/local/bin/kubectl 2>/dev/null || true
rm /tmp/oc.tar.gz
oc version --client
echo "oc CLI instalado correctamente."

# --- Docker ---
echo ""
echo "[4/4] Installing Docker..."
sudo apt-get update -q
sudo apt-get install -y -q docker.io
sudo usermod -aG docker "$USER"
echo "Docker instalado. IMPORTANTE: cierra y vuelve a abrir WSL2 para que tome efecto el grupo docker."

echo ""
echo "=================================================="
echo " Setup completado!"
echo " Ahora ejecuta: source ~/.bashrc"
echo " Luego ejecuta el script 2: ./2-connect-openshift.sh"
echo "=================================================="
