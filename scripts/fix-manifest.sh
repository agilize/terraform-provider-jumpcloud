#!/bin/bash
# Script para corrigir o arquivo de manifesto em um release existente

set -e

# Verifica se a versão foi fornecida
if [ $# -ne 1 ]; then
  echo "Uso: $0 <versão>"
  echo "Exemplo: $0 0.0.11"
  exit 1
fi

VERSION=$1
PROVIDER_NAME="terraform-provider-jumpcloud"
MANIFEST_FILE="${PROVIDER_NAME}_${VERSION}_manifest.json"
CHECKSUM_FILE="${PROVIDER_NAME}_${VERSION}_SHA256SUMS"
SIG_FILE="${CHECKSUM_FILE}.sig"

echo "🔨 Corrigindo o manifesto para a versão ${VERSION}"

# Criar diretório temporário
TMP_DIR=$(mktemp -d)
echo "📁 Diretório temporário: ${TMP_DIR}"

# Criar arquivo de manifesto
echo "📝 Criando arquivo de manifesto..."
cat > "${TMP_DIR}/${MANIFEST_FILE}" << EOF
{
  "version": 1,
  "metadata": {
    "protocol_versions": ["5.0", "6.0"]
  }
}
EOF

echo "✅ Arquivo de manifesto criado com sucesso"

# Baixar checksums existentes
if [ -f "${CHECKSUM_FILE}" ]; then
  echo "📋 Usando arquivo de checksum existente"
  cp "${CHECKSUM_FILE}" "${TMP_DIR}/"
else
  echo "🌐 Baixando checksums do GitHub..."
  curl -L -o "${TMP_DIR}/${CHECKSUM_FILE}" "https://github.com/agilize/terraform-provider-jumpcloud/releases/download/v${VERSION}/${CHECKSUM_FILE}"
fi

# Adicionar manifesto ao checksum
echo "🔐 Adicionando manifesto ao arquivo de checksums..."
(cd "${TMP_DIR}" && shasum -a 256 "${MANIFEST_FILE}" >> "${CHECKSUM_FILE}")

echo "📋 Conteúdo do arquivo de checksums:"
cat "${TMP_DIR}/${CHECKSUM_FILE}"

# Verificar se existe uma chave GPG
if gpg --list-secret-keys | grep -q "GPG"; then
  echo "🔑 Chave GPG encontrada, assinando o arquivo de checksums..."
  (cd "${TMP_DIR}" && gpg --detach-sign "${CHECKSUM_FILE}")
  
  echo "✅ Arquivo de checksums assinado com sucesso"
else
  echo "⚠️  Nenhuma chave GPG encontrada, não é possível assinar o arquivo de checksums"
  echo "⚠️  Você precisará assinar o arquivo manualmente com sua chave GPG registrada no Terraform Registry"
fi

# Copiar arquivos para o diretório atual
echo "📦 Copiando arquivos para o diretório atual..."
cp "${TMP_DIR}/${MANIFEST_FILE}" .
cp "${TMP_DIR}/${CHECKSUM_FILE}" .
if [ -f "${TMP_DIR}/${SIG_FILE}" ]; then
  cp "${TMP_DIR}/${SIG_FILE}" .
fi

echo ""
echo "🎉 Arquivos criados com sucesso!"
echo "📄 ${MANIFEST_FILE}"
echo "📄 ${CHECKSUM_FILE}"
if [ -f "${SIG_FILE}" ]; then
  echo "📄 ${SIG_FILE}"
fi

echo ""
echo "Para usar esses arquivos no GitHub Release, siga estas etapas:"
echo "1. Vá para https://github.com/agilize/terraform-provider-jumpcloud/releases/tag/v${VERSION}"
echo "2. Clique em 'Edit'"
echo "3. Faça upload dos arquivos gerados"
echo "4. Salve as alterações"

# Limpar
rm -rf "${TMP_DIR}"
echo "🧹 Diretório temporário removido"

echo ""
echo "⚠️  IMPORTANTE: Certifique-se de que o namespace em go.mod é 'registry.terraform.io/agilize/jumpcloud'"
echo "   e que os imports em main.go usam 'registry.terraform.io/agilize/jumpcloud/...'"
echo ""
echo "✨ Processo concluído! ✨" 