#!/bin/bash

# Função para corrigir o R014 em um arquivo
fix_r014() {
  local file=$1
  
  # Corrigir declarações de função e substituir o nome do parâmetro
  # Padrão: func xxxContextFunc(ctx context.Context, d *schema.ResourceData, m interface{})
  sed -i '' -E 's/func ([a-zA-Z0-9_]+)\(ctx context\.Context, d \*schema\.ResourceData, [a-zA-Z0-9_]+ interface{}\)/func \1(ctx context.Context, d *schema.ResourceData, meta interface{})/' "$file"
  
  # Corrigir referências ao parâmetro dentro da função (mais difícil)
  # Isso precisaria ser feito manualmente ou com uma ferramenta mais avançada,
  # pois exige análise de escopo e contexto.
  echo "Arquivo $file modificado, mas pode exigir ajustes manuais para referências internas."
}

# Encontrar todos os arquivos Go na pasta internal/provider
echo "Iniciando correção para R014..."
for file in $(find internal/provider -name "*.go"); do
  # Verificar se o arquivo contém possíveis erros R014
  if grep -q "func.*ctx context.Context.*d \*schema.ResourceData.*interface{}" "$file"; then
    echo "Processando $file..."
    fix_r014 "$file"
  fi
done

echo "Concluído. Verifique os arquivos alterados antes de fazer commit." 