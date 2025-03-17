# Configure the JumpCloud Provider
provider "jumpcloud" {
  api_key = var.jumpcloud_api_key # ou use variáveis de ambiente: JUMPCLOUD_API_KEY
  org_id  = var.jumpcloud_org_id  # ou use: JUMPCLOUD_ORG_ID
}

# Buscar política por nome
data "jumpcloud_policy" "password_policy" {
  name = "Secure Password Policy"
}

# Buscar política por ID
data "jumpcloud_policy" "mfa_policy" {
  id = "5f8b0e1b9d81b81b33c92a1c" # Exemplo de ID (substitua pelo ID real)
}

# Verificar se política está ativa
output "password_policy_status" {
  value = "${data.jumpcloud_policy.password_policy.name} está ${data.jumpcloud_policy.password_policy.active ? "ativa" : "inativa"}"
}

# Verificar configurações da política
output "password_policy_min_length" {
  value = lookup(data.jumpcloud_policy.password_policy.configurations, "min_length", "não especificado")
}

# Exibir tipo e template da política
output "mfa_policy_details" {
  value = "Política: ${data.jumpcloud_policy.mfa_policy.name}, Tipo: ${data.jumpcloud_policy.mfa_policy.type}, Criada em: ${data.jumpcloud_policy.mfa_policy.created}"
}

# Aplicar políticas condicionalmente
# Este exemplo verifica se uma política de MFA já existe, criando apenas se não existir
data "jumpcloud_policy" "existing_mfa" {
  name = "Required MFA Policy"
}

locals {
  mfa_policy_exists = data.jumpcloud_policy.existing_mfa.id != ""
}

resource "jumpcloud_policy" "conditional_mfa" {
  count       = local.mfa_policy_exists ? 0 : 1
  name        = "Required MFA Policy"
  description = "Política criada condicionalmente pelo Terraform"
  type        = "mfa"
  active      = true
  
  configurations = {
    require_mfa_for_all_users = "true"
  }
}

# Associar política existente a um grupo (requer recursos adicionais)
/*
resource "jumpcloud_user_group" "finance" {
  name = "Finance Department"
}

resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id  = data.jumpcloud_policy.password_policy.id
  group_id   = jumpcloud_user_group.finance.id
  type       = "user_group"
}
*/ 