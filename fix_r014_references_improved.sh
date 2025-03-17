#!/bin/bash

# Script para corrigir referências internas aos parâmetros 'm' renomeados para 'meta'
# Este script supõe que o primeiro script fix_r014.sh já foi executado para renomear os parâmetros nas assinaturas

echo "Iniciando correção de referências internas para parâmetros renomeados..."

# Iteramos sobre todos os arquivos Go no diretório provider que têm erros "undefined: m"
for file in $(grep -l "undefined: m" $(find internal/provider -name "*.go")); do
  echo "Processando arquivo: $file"
  
  # Criar arquivo temporário
  temp_file=$(mktemp)
  
  # Copiar o conteúdo original
  cat "$file" > "$temp_file"
  
  # Substituir todas as ocorrências de 'm.' por 'meta.' no arquivo inteiro
  sed -i '' "s/\bm\./meta./g" "$temp_file"
  
  # Substituir todas as ocorrências isoladas de 'm' por 'meta' no arquivo inteiro
  # (excluindo casos como 'time.m', 'fmt.m', etc.)
  sed -i '' "s/\([^a-zA-Z0-9_.]\)m\([^a-zA-Z0-9_]\)/\1meta\2/g" "$temp_file"
  
  # Aplicar as modificações de volta ao arquivo original
  mv "$temp_file" "$file"
  
  echo "  Corrigidas referências a 'm' no arquivo $file"
done

echo "Concluído! Verifique se todas as referências foram corretamente atualizadas." 