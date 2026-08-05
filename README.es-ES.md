

<div align="center">

# hush

**Ejecutor de comandos eficiente en contexto para agentes de código.**

Envuelve cualquier comando de shell e imprime una única línea de resumen `✓`/`✗`.\
En caso de éxito, muestra solo el resumen. En caso de fallo, muestra la salida filtrada. Conserva los códigos de salida.

[![CI](https://github.com/alfranz/hush/actions/workflows/ci.yml/badge.svg)](https://github.com/alfranz/hush/actions/workflows/ci.yml)
[![Release](https://github.com/alfranz/hush/actions/workflows/release.yml/badge.svg)](https://github.com/alfranz/hush/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alfranz/hush)](https://goreportcard.com/report/github.com/alfranz/hush)
[![GoDoc](https://pkg.go.dev/badge/github.com/alfranz/hush)](https://pkg.go.dev/github.com/alfranz/hush)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## Instalación

```bash
go install github.com/alfranz/hush/cmd/hush@latest
```

Esto coloca el binario `hush` en tu `$GOPATH/bin` (generalmente `~/go/bin`). Asegúrate de que ese directorio esté en tu `$PATH`:

```bash
# Verificar si ya está disponible
which hush

# Si no, añádelo a tu perfil de shell (~/.zshrc, ~/.bashrc, etc.)
export PATH="$HOME/go/bin:$PATH"
```

## Uso

```bash
# Uso básico
hush "pytest -x"
# ✓ pytest

# En caso de fallo, muestra la salida completa (rastros, detalles de aserciones, etc.)
hush "pytest -x"
# ✗ pytest
#   =================================== FAILURES ===================================
#   __________________________________ test_login __________________________________
#
#       def test_login():
#           response = client.get("/auth/login")
#   >       assert response.status_code == 200
#   E       AssertionError: assert 401 == 200
#
#   tests/test_auth.py:10: AssertionError
#   =========================== short test summary info ============================
#   FAILED tests/test_auth.py::test_login - AssertionError: assert 401 == 200
#   ============================== 1 failed in 0.06s ===============================

# Etiqueta personalizada
hush --label "unit tests" "pytest tests/unit"

# Filtrar salida en caso de fallo (úselo con precaución, véase la nota a continuación)
hush --tail 30 "npm test"          # últimas 30 líneas
hush --grep "error|FAIL" "make"    # líneas que coincidan con el patrón
hush --head 20 "cargo build"       # primeras 20 líneas

# Éxito con advertencias (el código de salida sigue siendo 0)
hush --warn-pattern "warning TS[0-9]+" --warn-tail 5 "tsc --noEmit"
# ⚠ tsc (3 warnings)
#   src/auth.ts:42:3 - warning TS2304: Cannot find name 'userId'.

# Modo por lotes
hush batch "ruff check ." "ty check src/" "pytest -x"
# ✓ ruff
# ✓ ty
# ✓ 2/2 checks passed

# Continuar en caso de fallo
hush batch --continue "ruff check ." "false" "pytest -x"
```

## Archivo de configuración (opcional)

Para comandos puntuales, los indicadores son suficientes. Un archivo de configuración es útil cuando tienes varias herramientas que ejecutar y deseas integrar los filtros adecuados para cada una, de modo que los agentes puedan simplemente llamar a `hush lint` o `hush all` sin repetir indicadores cada vez.

Crea `.hush.yaml` en la raíz de tu proyecto:

```yaml
defaults:
  continue: true

checks:
  lint:
    cmd: ruff check .
  types:
    cmd: ty check src/
    grep: "error:"
    warn-pattern: "warning TS[0-9]+"
    warn-tail: 5
  test:
    cmd: pytest -x
    tail: 40
```

Luego ejecuta las verificaciones nombradas:

```bash
hush lint          # ejecutar una sola verificación
hush all           # ejecutar todas las verificaciones en orden
```

La configuración en `defaults` se aplica a todos los comandos (raíz, por lotes y verificaciones nombradas) a menos que se anule mediante la configuración por verificación o indicadores de CLI. Precedencia: **Indicadores de CLI > configuración por verificación > defaults**.

## Indicadores

| Indicador | Descripción |
|------|-------------|
| `--label` | Etiqueta personalizada para la línea de resumen |
| `--tail N` | Mostrar solo las últimas N líneas en caso de fallo |
| `--head N` | Mostrar solo las primeras N líneas en caso de fallo |
| `--grep PATTERN` | Filtrar la salida a las líneas que coincidan |
| `--warn-pattern REGEX` | En caso de éxito, buscar líneas de advertencia y emitir `⚠` con detalles |
| `--warn-tail N` | En caso de éxito calificado con advertencias, mostrar las últimas N líneas de advertencia (predeterminado: 10) |
| `--continue` | Continuar ejecutando tras un fallo (por lotes/todos) |

> **Nota sobre `--grep` y fallos en pruebas:** Por defecto (sin indicadores), hush imprime la salida completa del comando en caso de fallo, incluyendo rastros, diferencias de aserciones y contexto del código fuente. Esto proporciona a los agentes la mayor cantidad de información para depurar. Utilice `--grep` y `--tail` principalmente para **herramientas de lint y compilación** que generen un gran volumen de salida. Para **ejecutores de pruebas** (pytest, Jest, go test), la salida sin filtrar suele ser lo que el agente necesita para corregir el problema. Un `--grep "FAIL"` en la salida de pytest, por ejemplo, elimina el rastro y los detalles de la aserción, dejando solo el resumen de una línea.

## Ahorro de tokens

El objetivo principal de hush es ahorrar tokens de contexto cuando los agentes de código ejecutan comandos de shell. Así es cómo se compara con los ejecutores de pruebas más populares:

**Pruebas exitosas** (salida de hush: única línea de resumen)

| Ejecutor de pruebas | Salida sin filtrar | Con hush | Ahorrado |
|---|---|---|---|
| pytest (47 pruebas) | ~1,028 tokens | ~4 tokens | **99%** |
| Jest (89 pruebas + cobertura) | ~886 tokens | ~5 tokens | **99%** |
| go test (14 paquetes) | ~196 tokens | ~6 tokens | **96%** |
| cargo test (51 pruebas) | ~683 tokens | ~5 tokens | **99%** |

**Pruebas fallidas** (salida de hush: resumen + contexto de error filtrado)

| Ejecutor de pruebas | Salida sin filtrar | Con hush | Ahorrado |
|---|---|---|---|
| pytest (1 fallo) | ~1,262 tokens | ~35 tokens | **97%** |
| Jest (1 fallo) | ~1,000 tokens | ~63 tokens | **93%** |
| go test (1 paquete fallido) | ~295 tokens | ~75 tokens | **74%** |

<details>
<summary><strong>Antes y después: pytest con 47 pruebas exitosas</strong></summary>

```
# Sin hush (50 líneas, ~1,028 tokens)
============================= test session starts ==============================
platform linux -- Python 3.12.1, pytest-8.0.2, pluggy-1.4.0
...
tests/test_auth.py::test_login_success PASSED                            [  2%]
tests/test_auth.py::test_login_invalid_password PASSED                   [  4%]
  ... (47 more lines)
=============================== 47 passed in 3.42s ================================

# Con hush (1 línea, ~4 tokens)
✓ pytest
```

</details>

Ejecute el benchmark completo usted mismo:

```bash
bash benchmark/run.sh
```

## Licencia

[MIT](LICENSE)
