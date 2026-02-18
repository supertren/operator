#!/bin/bash
# =============================================================================
# Script 3: Build and deploy hello-operator to OpenShift Local
# Run this script INSIDE Ubuntu WSL2, after running scripts 1 and 2
# =============================================================================

set -e

echo "=================================================="
echo " Hello Operator - Build & Deploy en OpenShift Local"
echo "=================================================="

# Ir al directorio del operador
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPERATOR_DIR="$SCRIPT_DIR/../hello-operator"
cd "$OPERATOR_DIR"

# --- Login en el registro interno de OpenShift ---
echo ""
echo "[1/5] Login en el registro interno de OpenShift..."
oc login -u kubeadmin https://api.crc.testing:6443 --insecure-skip-tls-verify

# Obtener token y URL del registro
OC_TOKEN=$(oc whoami -t)
REGISTRY="default-route-openshift-image-registry.apps-crc.testing"
IMG="${REGISTRY}/hello-operator-system/hello-operator:latest"

echo "Habilitando ruta del registro interno..."
oc patch configs.imageregistry.operator.openshift.io/cluster \
  --patch '{"spec":{"defaultRoute":true}}' \
  --type=merge 2>/dev/null || true
sleep 3

echo "Login en registro Docker..."
echo "$OC_TOKEN" | docker login "$REGISTRY" -u kubeadmin --password-stdin

# --- Crear namespace ---
echo ""
echo "[2/5] Creando namespace hello-operator-system..."
oc new-project hello-operator-system 2>/dev/null || oc project hello-operator-system

# --- Generar go.sum y descargar dependencias ---
echo ""
echo "[3/5] Generando go.sum y descargando dependencias Go..."
go mod tidy
go mod download

# --- Build y push de la imagen ---
echo ""
echo "[4/5] Build de la imagen Docker..."
docker build -t "$IMG" .

echo "Push de la imagen al registro de OpenShift..."
docker push "$IMG"

# --- Desplegar el operador ---
echo ""
echo "[5/5] Desplegando el operador en OpenShift..."

# CRDs y ClusterRoles requieren cluster-admin (kubeadmin)
echo "Instalando CRD y RBAC de cluster (requiere kubeadmin)..."
oc login -u kubeadmin https://api.crc.testing:6443 --insecure-skip-tls-verify
oc project hello-operator-system
oc apply -f config/crd/bases/
oc apply -f config/rbac/service_account.yaml
oc apply -f config/rbac/role.yaml
oc apply -f config/rbac/role_binding.yaml

# Namespace y recursos namespace-level con developer
echo "Desplegando manager (developer)..."
oc login -u developer -p developer https://api.crc.testing:6443 --insecure-skip-tls-verify
oc project hello-operator-system
sed "s|controller:latest|${IMG}|g" config/manager/manager.yaml | oc apply -f -

# Esperar a que el deployment esté listo
echo ""
echo "Esperando a que el operador esté listo..."
oc rollout status deployment/hello-operator-controller-manager -n hello-operator-system --timeout=120s

echo ""
echo "=================================================="
echo " Operador desplegado!"
echo ""
echo " Aplicando CR de ejemplo..."
oc apply -f config/samples/hello_v1alpha1_helloapp.yaml

echo ""
echo " Estado del operador:"
oc get deployment -n hello-operator-system
echo ""
echo " CRs HelloApp:"
oc get helloapp -n hello-operator-system
echo ""
echo " Para ver los logs del operador:"
echo "   oc logs -l control-plane=controller-manager -n hello-operator-system -f"
echo "=================================================="
