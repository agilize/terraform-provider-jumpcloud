# jumpcloud_radius_server Data Source

Este data source permite obter informações sobre um servidor RADIUS específico configurado no JumpCloud. Pode ser útil para referenciar servidores RADIUS existentes sem precisar recriá-los em seu código Terraform.

## Exemplo de Uso

### Buscar por ID

```hcl
data "jumpcloud_radius_server" "vpn_server" {
  id = "5f8d3e9c9d1d8b6a8c4c2f7g"
}

output "vpn_server_mfa_required" {
  value = data.jumpcloud_radius_server.vpn_server.mfa_required
}
```

### Buscar por Nome

```hcl
data "jumpcloud_radius_server" "wifi_auth" {
  name = "WiFi Authentication"
}

# Usar o ID em outro recurso ou associação
resource "jumpcloud_user_group" "wifi_users" {
  name        = "WiFi Users"
  description = "Usuários com acesso à rede WiFi autenticada"
}

resource "jumpcloud_radius_server" "new_wifi_radius" {
  name          = "New WiFi Authentication"
  shared_secret = var.radius_secret
  
  # Associar com o mesmo grupo do servidor existente
  targets = [
    jumpcloud_user_group.wifi_users.id
  ]
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `id` - (Opcional) ID do servidor RADIUS no JumpCloud. Conflita com `name`.
* `name` - (Opcional) Nome do servidor RADIUS no JumpCloud. Conflita com `id`.

**Nota**: Exatamente um de `id` ou `name` deve ser especificado.

## Attribute Reference

Os seguintes atributos são exportados:

* `network_source_ip` - IP de origem da rede usado para comunicação com o servidor RADIUS.
* `mfa_required` - Se a autenticação multifator é exigida para o servidor RADIUS.
* `user_password_expiration_action` - Ação a ser tomada quando a senha do usuário expirar (`allow` ou `deny`).
* `user_lockout_action` - Ação a ser tomada quando o usuário for bloqueado (`allow` ou `deny`).
* `user_attribute` - Atributo do usuário usado para autenticação (`username` ou `email`).
* `targets` - Lista de IDs de grupos associados ao servidor RADIUS.
* `created` - Data de criação do servidor RADIUS.
* `updated` - Data da última atualização do servidor RADIUS.

**Nota**: O atributo `shared_secret` não é exportado por razões de segurança. 