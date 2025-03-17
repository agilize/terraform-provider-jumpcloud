# jumpcloud_radius_server Resource

Este recurso permite gerenciar servidores RADIUS no JumpCloud. O RADIUS (Remote Authentication Dial-In User Service) é um protocolo de rede que fornece autenticação, autorização e contabilização centralizada para usuários que se conectam e usam serviços de rede.

## Exemplo de Uso

### Configuração básica de servidor RADIUS

```hcl
resource "jumpcloud_radius_server" "corporate_vpn" {
  name          = "Corporate VPN"
  shared_secret = "s3cur3-sh4r3d-s3cr3t"
  mfa_required  = true
  
  # IP de origem para comunicação com o servidor RADIUS
  network_source_ip = "10.0.1.5"
  
  # Configurações de autenticação
  user_attribute = "username"
  user_password_expiration_action = "deny"
  user_lockout_action = "deny"
}
```

### Servidor RADIUS associado a grupos de usuários

```hcl
resource "jumpcloud_radius_server" "wifi_auth" {
  name          = "WiFi Authentication"
  shared_secret = var.radius_secret
  mfa_required  = false
  
  # Configurar autenticação por e-mail ao invés de nome de usuário
  user_attribute = "email"
  
  # Associar com grupos específicos
  targets = [
    jumpcloud_user_group.employees.id,
    jumpcloud_user_group.contractors.id
  ]
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `name` - (Obrigatório) Nome do servidor RADIUS.
* `shared_secret` - (Obrigatório) Segredo compartilhado usado para autenticação entre cliente e servidor RADIUS. Este valor é sensível e não será exibido na saída do Terraform.
* `network_source_ip` - (Opcional) IP de origem da rede que será usada para se comunicar com o servidor RADIUS.
* `mfa_required` - (Opcional) Se a autenticação multifator é exigida para o servidor RADIUS. Padrão: `false`.
* `user_password_expiration_action` - (Opcional) Ação a ser tomada quando a senha do usuário expirar. Valores válidos: `allow` ou `deny`. Padrão: `allow`.
* `user_lockout_action` - (Opcional) Ação a ser tomada quando o usuário for bloqueado. Valores válidos: `allow` ou `deny`. Padrão: `deny`.
* `user_attribute` - (Opcional) Atributo do usuário usado para autenticação. Valores válidos: `username` ou `email`. Padrão: `username`.
* `targets` - (Opcional) Lista de IDs de grupos de usuários associados ao servidor RADIUS. Se não especificado, o servidor estará disponível para todos os usuários.

## Attribute Reference

Além dos argumentos listados acima, os seguintes atributos são exportados:

* `id` - ID do servidor RADIUS.
* `created` - Data de criação do servidor RADIUS.
* `updated` - Data da última atualização do servidor RADIUS.

## Import

Servidores RADIUS JumpCloud podem ser importados usando o ID do servidor:

```
terraform import jumpcloud_radius_server.example {radius_server_id}
``` 