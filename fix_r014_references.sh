#!/bin/bash

# Função para corrigir as referências internas em um arquivo
fix_parameter_references() {
  local file=$1
  local line_start=$2
  local param_name=$3

  # Extrair o conteúdo da função
  function_start=$(grep -n "func.*ctx context.Context.*meta interface{}" "$file" | awk -F':' '{print $1}')
  
  if [ -z "$function_start" ]; then
    echo "Não foi possível encontrar funções a serem processadas em $file"
    return
  fi
  
  # Para cada função encontrada
  for line_num in $function_start; do
    # Encontrar o bloco de função
    # Isso é uma aproximação, pois precisaríamos de um parser real para analisar o escopo
    next_func=$(tail -n +$((line_num+1)) "$file" | grep -n "^func" | head -1 | awk -F':' '{print $1}')
    
    if [ -z "$next_func" ]; then
      # Se não houver próxima função, vamos até o final do arquivo
      block_end=$(wc -l "$file" | awk '{print $1}')
    else
      block_end=$((line_num + next_func - 1))
    fi
    
    # Extrair o bloco atual
    block_start=$((line_num))
    block=$(sed -n "${block_start},${block_end}p" "$file")
    
    # Analisar o nome do parâmetro na declaração da função
    func_decl=$(echo "$block" | head -1)
    old_param=$(echo "$func_decl" | grep -o -E '\b[a-zA-Z0-9_]+\s+interface\{\}' | awk '{print $1}')
    
    if [ -z "$old_param" ] || [ "$old_param" = "meta" ]; then
      # Se já for meta ou não encontrarmos o parâmetro, pular
      continue
    fi
    
    echo "Processando função na linha $line_num, substituindo parâmetro '$old_param' por 'meta'"
    
    # Criar um arquivo temporário para essa função específica
    temp_file=$(mktemp)
    
    # Substituir o parâmetro dentro do corpo da função
    if [ ! -z "$old_param" ]; then
      # Extrair e modificar o bloco de função
      sed -n "${block_start},${block_end}p" "$file" | sed -E "s/\b${old_param}([^a-zA-Z0-9_])/meta\1/g" > "$temp_file"
      
      # Substituir o bloco original pelo modificado
      sed -i '' "${block_start},${block_end}d" "$file"
      sed -i '' "${block_start}r $temp_file" "$file"
    fi
    
    rm "$temp_file"
  done
}

# Encontrar todos os arquivos Go modificados pelo primeiro script
echo "Iniciando correção de referências internas..."
for file in $(find internal/provider -name "*.go"); do
  if grep -q "func.*ctx context.Context.*meta interface{}" "$file"; then
    echo "Processando referências em $file..."
    fix_parameter_references "$file"
  fi
done

echo "Concluído. Verifique as alterações antes de fazer commit." 