#!/bin/bash

# Script para corrigir problemas de linting relacionados à regra R014 do tfproviderlint
# Esta regra requer que parâmetros do tipo interface{} sejam chamados de 'meta'

echo "Iniciando correção completa para regra R014..."

# Encontrar todos os arquivos Go no diretório provider
for file in $(find internal/provider -name "*.go"); do
  # Verificar se o arquivo contém funções com parâmetros interface{}
  if grep -q "func.*interface{}" "$file"; then
    echo "Processando arquivo: $file"
    
    # Criar arquivo temporário
    temp_file=$(mktemp)
    
    # Primeiro, renomeamos os parâmetros nas assinaturas de função
    # Substituímos qualquer nome de parâmetro seguido de "interface{}" para "meta interface{}"
    sed 's/\([a-zA-Z0-9_]\+\) interface{}/meta interface{}/g' "$file" > "$temp_file"
    
    # Procuramos por padrões que indicam uso do parâmetro antigo
    # Esta etapa é mais complexa e pode exigir ajustes
    # Procuramos por m. e m como token isolado
    sed -i '' 's/\bm\./meta./g' "$temp_file"
    sed -i '' 's/\([^a-zA-Z0-9_.]\)m\([^a-zA-Z0-9_]\)/\1meta\2/g' "$temp_file"
    
    # Casos especiais para cuidar do 'm' no início ou final de uma linha
    sed -i '' 's/^m\([^a-zA-Z0-9_]\)/meta\1/g' "$temp_file"
    sed -i '' 's/\([^a-zA-Z0-9_.]\)m$/\1meta/g' "$temp_file"
    
    # Aplicar as modificações de volta ao arquivo original
    mv "$temp_file" "$file"
    
    echo "  Corrigidos parâmetros e referências no arquivo $file"
  fi
done

echo "Concluído! Verifique se todas as correções foram realizadas adequadamente." 