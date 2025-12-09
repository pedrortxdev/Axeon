#!/bin/bash

echo "🚀 Iniciando o pré-carregamento (Bypass do Bloqueio LXD)..."

# Função para baixar da fonte OFICIAL (ubuntu:) e salvar com nosso alias interno
download_official() {
    REMOTE_SOURCE=$1  # ex: ubuntu:22.04 (Fonte Oficial da Canonical)
    LOCAL_ALIAS=$2    # ex: ubuntu/22.04 (O nome que nosso Frontend espera)

    echo "------------------------------------------------"
    echo "⬇️  Baixando: $REMOTE_SOURCE -> Local: $LOCAL_ALIAS"
    
    # Copia do remote 'ubuntu:' (que não está bloqueado) para 'local:'
    lxc image copy $REMOTE_SOURCE local: --alias $LOCAL_ALIAS --auto-update --public
}

# 1. Ubuntu: Usamos o remote 'ubuntu:' (Canonical)
download_official "ubuntu:22.04" "ubuntu/22.04"
# download_official "ubuntu:24.04" "ubuntu/24.04" # Se quiser o mais novo

# 2. Alpine/Debian: O 'images:' está bloqueado para LXD. 
# Truque: Vamos tentar baixar o Alpine via rootfs ou ignorar por enquanto e focar no Ubuntu.
# Se você realmente precisar de Alpine, teremos que importar o tarball manualmente.
# Por enquanto, vamos focar no que funciona: Ubuntu.

echo "------------------------------------------------"
echo "✅ Imagens Ubuntu baixadas."
echo "⚠️  Nota: Alpine/Debian via 'images:' estão bloqueados para LXD."
echo "📋 Listando imagens locais:"
lxc image list
