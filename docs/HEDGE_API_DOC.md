# API Documentation: Hedge Strategy (Dual-Wallet)

Esta documentación detalla los endpoints necesarios para integrar la funcionalidad de **Hedge Strategy** en el frontend.

## Base URL

`http://localhost:8081/api`

## Autenticación

Todos los endpoints requieren un token JWT en el header `Authorization`.

```http
Authorization: Bearer <your_access_token>
```

---

## Endpoints

### 1. Obtener Detalles de la Estrategia

**GET** `/hedge/strategy`

Obtiene la información general de la estrategia activa.

- **Response (200 OK):**

```json
{
  "strategy_id": "TW-8892",
  "name": "Estrategia Dual-Wallet (LP + Hedge)",
  "description": "Separe su capital de inversión del capital de cobertura con seguridad aislada.",
  "status": "active",
  "is_synced": true,
  "created_at": "2024-01-15T10:00:00Z"
}
```

### 2. Obtener KPIs (Stats)

**GET** `/hedge/stats`

Obtiene las métricas principales para los dashboards.

- **Response (200 OK):**

```json
{
  "apr": { "value": 113.82, "trend": 12.4, "trend_direction": "up" },
  "fees_accumulated": { "value": 10.71, "currency": "USD", "trend": 5.2, "trend_direction": "up" },
  "delta": { "value": 0.0, "unit": "ETH" },
  "hedge_efficiency": { "value": 99.8, "trend": 0.2, "trend_direction": "up" }
}
```

### 3. Estado de Wallets

**GET** `/hedge/wallets`

Estado de conexión de Wallet A (LP) y Wallet B (Hedge).

- **Response (200 OK):**

```json
{
  "wallet_a": {
    "type": "uniswap_lp",
    "name": "Wallet A (Uniswap LP)",
    "description": "Gestiona su posición en Uniswap V3 Arbitrum via Permit2.",
    "connected": true,
    "address": "0x71C...4921",
    "full_address": "0x71C7656ec7ab88b098defB751B7401B5f6d4921",
    "permissions": ["add_liquidity", "collect_fees"],
    "permissions_note": "Permisos de solo gestión de liquidez aprobados"
  },
  "wallet_b": {
    "type": "hyperliquid_trade",
    "name": "Wallet B (Hyperliquid Trade)",
    "description": "Dedicada a la cobertura (Short). Utiliza un Signing Agent.",
    "connected": false,
    "address": null,
    "full_address": null,
    "permissions": ["adjust_short"],
    "permissions_note": "El bot no tiene permisos de retiro"
  }
}
```

### 4. Conectar Wallet

**POST** `/hedge/wallets/connect`

- **Body:**

```json
{
  "wallet_type": "wallet_a",
  "address": "0x71C7656ec7ab88b098defB751B7401B5f6d4921"
}
```

### 5. Sincronización de Delta (Live)

**GET** `/hedge/delta`

*Nota: Se recomienda polling corto o actualización vía WebSocket en el futuro.*

- **Response (200 OK):**

```json
{
  "pool_exposure": { "value": 0.2522, "unit": "WETH", "percentage": 65 },
  "hedge_exposure": { "value": -0.2522, "unit": "WETH", "percentage": 65 },
  "net_exposure": 0.0,
  "status": "neutral",
  "is_live": true,
  "last_sync": "2024-01-15T10:05:00Z"
}
```

### 6. Matriz de Permisos

**GET** `/hedge/permissions`

- **Response (200 OK):**

```json
{
  "permissions": [
    {
      "action": "Añadir Liquidez",
      "wallet_a": { "required": "Firma Requerida", "type": "user_signature" },
      "wallet_b": { "required": "No aplica", "type": "not_applicable" }
    }
  ]
}
```

### 7. Modo Seguro

**GET** `/hedge/safe-mode`

### 8. Flujo Técnico de Sincronización

**GET** `/hedge/sync-flow`

---

## Sugerencias para el Equipo Frontend (Adapting UI)

1. **Gestión de Estados**: Utilizar los campos `connected` de `/wallets` para habilitar/deshabilitar botones de rebalanceo o acciones de LP.
2. **Alertas de Sincronización**: Mostrar el estado de `net_exposure` de `/delta`. Si es `!= 0.0`, resaltar en amarillo como "Sincronizando".
3. **Seguridad**: El campo `permissions_note` debe mostrarse cerca de la conexión de la wallet para dar tranquilidad al usuario de que el bot no tiene permisos de retiro.
4. **Modo Seguro**: Si `is_active` en `/safe-mode` es `true`, bloquear todas las acciones de trading y mostrar banner rojo con el `trigger_reason`.
