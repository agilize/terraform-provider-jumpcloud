# Exemplos do Data Source jumpcloud_system

Este diretório contém exemplos de uso do data source `jumpcloud_system` do Provider Terraform para JumpCloud.

## Visão Geral

O data source `jumpcloud_system` permite recuperar informações sobre sistemas (dispositivos) existentes no JumpCloud através do Terraform. Isso é particularmente útil para referenciar sistemas existentes em novas configurações ou para criar relatórios e dashboards com informações dos sistemas.

## Exemplos Disponíveis

### data-source.tf

Exemplo completo de consulta de sistemas JumpCloud, incluindo:
- Busca de sistemas por nome de exibição
- Busca de sistemas por ID
- Integração com outros recursos
- Uso de filtragem e lógica condicional
- Exemplo de integração com ferramentas de monitoramento
- Outputs organizados por categorias

## Como Usar Estes Exemplos

1. **Prepare seu ambiente:**
   ```bash
   # Configure as variáveis de ambiente para autenticação
   export JUMPCLOUD_API_KEY="sua-api-key"
   export JUMPCLOUD_ORG_ID="seu-org-id"
   ```

2. **Inicialize o Terraform:**
   ```bash
   terraform init
   ```

3. **Adapte os exemplos** às suas necessidades:
   - Substitua os nomes de exibição e IDs pelos valores reais dos seus sistemas
   - Modifique os outputs conforme necessário para seu caso de uso
   - Ajuste as condições de filtragem para corresponder às suas necessidades

4. **Execute uma consulta de teste:**
   ```bash
   terraform plan
   ```

## Casos de Uso Comuns

1. **Inventário e Relatórios:**
   - Gerar inventários detalhados de sistemas
   - Criar dashboards com métricas de segurança
   - Monitorar configurações de sistemas

2. **Referências para Outros Recursos:**
   - Associar usuários a sistemas existentes
   - Associar sistemas a grupos
   - Aplicar políticas a sistemas específicos

3. **Automação e Workflows:**
   - Usar informações dos sistemas em scripts e automações
   - Implementar validações e verificações de conformidade
   - Detectar sistemas com configurações inadequadas

## Combinando com Outros Recursos

Estes exemplos de data sources funcionam melhor quando combinados com:
- Recursos `jumpcloud_system` para gerenciar configurações
- Recursos `jumpcloud_user` para gerenciar usuários
- Recursos de associação para conectar sistemas a usuários e grupos
- Outros data sources para buscar informações complementares

## Dicas e Melhores Práticas

1. **Performance:**
   - Evite buscar todos os sistemas quando possível, use filtros precisos
   - Mantenha o número de consultas ao mínimo necessário

2. **Organização:**
   - Use outputs estruturados para organizar informações
   - Combine data sources com locals para processamento intermediário

3. **Segurança:**
   - Evite exibir informações sensíveis em outputs não marcados como sensitive
   - Use os data sources para validar configurações de segurança

## Recursos Relacionados

- Documentação do data source: `terraform-provider-jumpcloud/docs/data-sources/system.md`
- API JumpCloud: [https://docs.jumpcloud.com/api/](https://docs.jumpcloud.com/api/)
- Documentação do JumpCloud: [https://support.jumpcloud.com/](https://support.jumpcloud.com/) 