terraform {
  required_providers {
    jumpcloud = {
      source  = "registry.terraform.io/agilize/jumpcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "jumpcloud" {
  api_key = var.jumpcloud_api_key
}

resource "jumpcloud_user" "example" {
  username  = "example.user"
  email     = "example.user@example.com"
  firstname = "Example"
  lastname  = "User"
  
  password = "SecurePassword123!"
  
  # Atributos adicionais
  employeeType    = "contractor"
  jobTitle        = "DevOps Engineer"
  department      = "IT"
  costCenter      = "CC-123"
  company         = "Example Inc."
  description     = "Conta de usuário para exemplo"
  location        = "Remote"
  
  # Atributos personalizados/custom attributes
  attributes = {
    customAttribute1 = "valor1"
    customAttribute2 = "valor2"
  }
  
  # Tags para organização
  tags = ["dev", "terraform-managed"]
  
  # Configuração de MFA
  mfa = {
    configured = false
    exclusion  = true
    exclusion_until = "2023-12-31"
  }
}

# Exemplo de ativação/desativação de usuário
resource "jumpcloud_user" "temporary" {
  username  = "temp.user"
  email     = "temp.user@example.com"
  firstname = "Temporary"
  lastname  = "User"
  
  password = "SecurePassword456!"
  
  # Usuário será criado como inativo
  state = "DISABLED"
  
  lifecycle {
    ignore_changes = [
      # Ignorar mudanças feitas ao password após a criação
      password,
    ]
  }
} 