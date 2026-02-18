#!/bin/bash
# =============================================================================
# Script 2: Connect WSL2 to OpenShift Local (CRC) running on Windows
# Run this script INSIDE Ubuntu WSL2
# =============================================================================

set -e

echo "=================================================="
echo " Conectando WSL2 a OpenShift Local (CRC)"
echo "=================================================="

# Obtener la IP del host Windows desde WSL2
WINDOWS_IP=$(ip route | grep default | awk '{print $3}')
echo ""
echo "IP de Windows detectada: $WINDOWS_IP"

# Añadir api.crc.testing al /etc/hosts de WSL2
if grep -q "api.crc.testing" /etc/hosts; then
  echo "api.crc.testing ya existe en /etc/hosts, actualizando..."
  sudo sed -i "/api.crc.testing/d" /etc/hosts
fi
echo "$WINDOWS_IP api.crc.testing" | sudo tee -a /etc/hosts
echo "$WINDOWS_IP console-openshift-console.apps-crc.testing" | sudo tee -a /etc/hosts
echo "$WINDOWS_IP oauth-openshift.apps-crc.testing" | sudo tee -a /etc/hosts
echo "$WINDOWS_IP default-route-openshift-image-registry.apps-crc.testing" | sudo tee -a /etc/hosts
echo ""
echo "Entradas añadidas a /etc/hosts"

# Copiar el certificado de CRC desde Windows (si existe)
WIN_USER=$(cmd.exe /c 'echo %USERNAME%' 2>/dev/null | tr -d '\r')
CRC_CERT_PATH="/mnt/c/Users/${WIN_USER}/.crc/machines/crc/kubeconfig"
if [ -f "$CRC_CERT_PATH" ]; then
  mkdir -p ~/.kube
  cp "$CRC_CERT_PATH" ~/.kube/config
  echo "kubeconfig copiado desde CRC."
else
  echo ""
  echo "AVISO: No se encontró el kubeconfig de CRC en $CRC_CERT_PATH"
  echo "Intenta hacer login manualmente:"
fi

echo ""
echo "Intentando login en OpenShift Local..."
echo "Si falla, ejecuta manualmente:"
echo "  oc login -u developer -p developer https://api.crc.testing:6443 --insecure-skip-tls-verify"
echo ""

oc login -u developer -p developer https://api.crc.testing:6443 --insecure-skip-tls-verify || {
  echo ""
  echo "Login automático fallido. Por favor ejecuta el login manualmente (ver instrucciones arriba)."
  echo "Asegúrate de que OpenShift Local esté corriendo en Windows (crc start)"
  exit 1
}

echo ""
echo "=================================================="
echo " Conectado a OpenShift Local!"
oc get nodes
echo ""
echo " Ahora ejecuta el script 3: ./3-build-and-deploy.sh"
echo "=================================================="
