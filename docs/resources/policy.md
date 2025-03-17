# jumpcloud_policy Resource

Este recurso permite gerenciar políticas no JumpCloud. As políticas são configurações que podem ser aplicadas a usuários ou sistemas, controlando diferentes aspectos de segurança como complexidade de senha, MFA, bloqueio de conta e atualizações de sistema.

## Exemplo de Uso

```hcl
# Exemplo de política de complexidade de senha
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Política de complexidade de senha segura para todos os usuários"
  type        = "password_complexity"
  active      = true
  
  # Configurações específicas para política de complexidade de senha
  configurations = {
    min_length             = "12"         # Comprimento mínimo da senha
    requires_uppercase     = "true"       # Requer letras maiúsculas
    requires_lowercase     = "true"       # Requer letras minúsculas
    requires_number        = "true"       # Requer números
    requires_special_char  = "true"       # Requer caracteres especiais
    password_expires_days  = "90"         # Senha expira em 90 dias
    enable_password_expiry = "true"       # Ativar expiração de senha
  }
}

# Exemplo de política de MFA (Autenticação Multi-Fator)
resource "jumpcloud_policy" "mfa_policy" {
  name        = "Required MFA Policy"
  description = "Política que requer MFA para todos os usuários"
  type        = "mfa"
  active      = true
  
  configurations = {
    allow_sms_enrollment        = "true"   # Permitir MFA via SMS
    allow_voice_call_enrollment = "true"   # Permitir MFA via chamada de voz
    allow_totp_enrollment       = "true"   # Permitir MFA via aplicativo (TOTP)
    allow_push_notification     = "true"   # Permitir MFA via notificação push
    require_mfa_for_all_users   = "true"   # Exigir MFA para todos os usuários
  }
}
```

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `name` - (Obrigatório) Nome da política.
* `type` - (Obrigatório) Tipo da política. Valores válidos: `password_complexity`, `samba_ad_password_sync`, `password_expiration`, `custom`, `password_reused`, `password_failed_attempts`, `account_lockout_timeout`, `mfa`, `system_updates`.
* `active` - (Opcional) Indica se a política está ativa. Padrão: `true`.
* `description` - (Opcional) Descrição da política.
* `template` - (Opcional) Template a ser usado pela política. Necessário para alguns tipos de política.
* `configurations` - (Opcional) Mapa de configurações específicas para o tipo de política. Cada tipo de política requer diferentes configurações.

## Tipos de Políticas e Configurações

### Complexidade de Senha (`password_complexity`)

Configurações disponíveis:
* `min_length` - Comprimento mínimo da senha
* `requires_uppercase` - Se requer letras maiúsculas ("true"/"false") 
* `requires_lowercase` - Se requer letras minúsculas ("true"/"false")
* `requires_number` - Se requer números ("true"/"false") 
* `requires_special_char` - Se requer caracteres especiais ("true"/"false")
* `password_expires_days` - Após quantos dias a senha expira
* `enable_password_expiry` - Se ativa a expiração de senha ("true"/"false")

### MFA (`mfa`)

Configurações disponíveis:
* `allow_sms_enrollment` - Permite MFA via SMS ("true"/"false")
* `allow_voice_call_enrollment` - Permite MFA via chamada de voz ("true"/"false")
* `allow_totp_enrollment` - Permite MFA via aplicativo autenticador (TOTP) ("true"/"false")
* `allow_push_notification` - Permite MFA via notificação push ("true"/"false")
* `require_mfa_for_all_users` - Exige MFA para todos os usuários ("true"/"false")
* `days_before_password_expiration` - Dias antes da expiração para lembrar o usuário

### Bloqueio de Conta (`account_lockout_timeout`)

Configurações disponíveis:
* `max_failed_login_attempts` - Número máximo de tentativas de login falhas
* `lockout_time_in_minutes` - Tempo de bloqueio em minutos
* `reset_counter_after_minutes` - Minutos para resetar o contador após período de inatividade

### Atualizações de Sistema (`system_updates`)

Configurações disponíveis:
* `auto_update_enabled` - Ativa atualizações automáticas ("true"/"false")
* `auto_update_time` - Horário para realizar as atualizações (formato "HH:MM")
* `notify_users` - Notifica usuários sobre atualizações pendentes ("true"/"false")

## Referência de Atributos

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - O ID da política.
* `created` - A data de criação da política.

## Importação

Políticas podem ser importadas usando o ID da política:

```
terraform import jumpcloud_policy.example 5f0c1b2c3d4e5f6g7h8i9j0k
```

## Relacionamentos com Outros Recursos

As políticas podem ser associadas a grupos de usuários ou sistemas usando o recurso `jumpcloud_policy_association`:

```hcl
resource "jumpcloud_user_group" "finance" {
  name = "Finance Department"
}

resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id = jumpcloud_policy.password_complexity.id
  group_id  = jumpcloud_user_group.finance.id
  type      = "user_group"
}
```

## Considerações de Segurança

* Recomenda-se definir políticas de senha fortes, especialmente para ambientes de produção.
* Para políticas de MFA, considere o equilíbrio entre segurança e usabilidade ao escolher os métodos permitidos.
* Para sistemas críticos, é aconselhável usar políticas mais restritivas.
* Monitore regularmente a conformidade das políticas para garantir que estão sendo aplicadas corretamente. 