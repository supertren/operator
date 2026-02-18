# Hello Operator

Operador básico de OpenShift creado con operator-sdk y Go.

## Qué hace este operador

Cuando creas un recurso `HelloApp`, el operador crea automáticamente un **Deployment** de nginx con el mensaje que especifiques.

```yaml
apiVersion: hello.example.com/v1alpha1
kind: HelloApp
metadata:
  name: mi-app
spec:
  message: "Hola desde OpenShift!"
  replicas: 2
```

## Requisitos previos

- Windows 11 con OpenShift Local (CRC) corriendo
- WSL2 con Ubuntu instalado

## Instalación paso a paso

### Paso 1 — Instalar WSL2 (PowerShell como Administrador en Windows)

```powershell
wsl --install -d Ubuntu
```

Reinicia el PC cuando termine.

### Paso 2 — Clonar el repo en WSL2

Abre Ubuntu (WSL2) y ejecuta:

```bash
git clone git@github.com:supertren/operator.git
cd operator
```

### Paso 3 — Instalar herramientas en WSL2

```bash
chmod +x scripts/*.sh
./scripts/1-setup-wsl2.sh
source ~/.bashrc
```

### Paso 4 — Conectar WSL2 a OpenShift Local

Asegúrate de que OpenShift Local esté corriendo en Windows:

```powershell
# En PowerShell de Windows:
crc start
```

Luego en WSL2:

```bash
./scripts/2-connect-openshift.sh
```

### Paso 5 — Build y deploy del operador

```bash
./scripts/3-build-and-deploy.sh
```

### Paso 6 — Verificar que funciona

```bash
# Ver el operador corriendo
oc get pods -n hello-operator-system

# Ver el HelloApp creado
oc get helloapp -n hello-operator-system

# Ver el Deployment que creó el operador
oc get deployment -n hello-operator-system

# Ver logs del operador
oc logs -l control-plane=controller-manager -n hello-operator-system -f
```

## Estructura del proyecto

```
.
├── hello-operator/
│   ├── api/v1alpha1/          # Tipos del CRD (HelloApp)
│   ├── internal/controller/   # Lógica del controlador
│   ├── config/
│   │   ├── crd/bases/         # CRD YAML generado
│   │   ├── rbac/              # ServiceAccount, Role, RoleBinding
│   │   ├── manager/           # Deployment del operador
│   │   └── samples/           # CR de ejemplo
│   ├── Dockerfile
│   ├── Makefile
│   ├── main.go
│   └── go.mod
└── scripts/
    ├── 1-setup-wsl2.sh        # Instala Go, operator-sdk, oc, Docker en WSL2
    ├── 2-connect-openshift.sh # Conecta WSL2 a OpenShift Local
    └── 3-build-and-deploy.sh  # Build, push y deploy del operador
```

## Comandos útiles

```bash
# Eliminar el CR de ejemplo
oc delete -f hello-operator/config/samples/hello_v1alpha1_helloapp.yaml

# Desinstalar el operador
cd hello-operator && make undeploy && make uninstall

# Crear un nuevo HelloApp personalizado
oc apply -f - <<EOF
apiVersion: hello.example.com/v1alpha1
kind: HelloApp
metadata:
  name: mi-helloapp
  namespace: hello-operator-system
spec:
  message: "Mi mensaje personalizado"
  replicas: 3
EOF
```
