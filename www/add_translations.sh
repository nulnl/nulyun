#!/bin/bash
# Add storage quota translations to all remaining i18n files

cd /Users/dukangxu/dev/nul/nulyun/www/src/i18n

# Italian
cat > /tmp/it_patch.json << 'EOF'
{
  "errors": {
    "quotaExceeded": "Quota di archiviazione superata. Eliminare alcuni file o contattare l'amministratore.",
    "quotaExceededDuringUpload": "Caricamento interrotto: quota di archiviazione superata. Il file parziale è stato eliminato.",
    "insufficientSpace": "Spazio di archiviazione insufficiente per questo caricamento."
  },
  "settings": {
    "storageQuota": "Quota di Archiviazione",
    "storageQuotaHelp": "Imposta lo spazio di archiviazione massimo per questo utente. Usa formati come 10M, 5G, ecc. Imposta su 0 o lascia vuoto per illimitato (solo amministratore).",
    "unlimited": "Illimitato"
  },
  "sidebar": {
    "storageUsed": "di {total} utilizzato",
    "storageUsedUnlimited": "utilizzato (Illimitato)"
  }
}
EOF

# Japanese
cat > /tmp/ja_patch.json << 'EOF'
{
  "errors": {
    "quotaExceeded": "ストレージ容量を超えました。ファイルを削除するか、管理者に連絡してください。",
    "quotaExceededDuringUpload": "アップロード停止：ストレージ容量を超えました。部分ファイルは削除されました。",
    "insufficientSpace": "このアップロードに十分なストレージ容量がありません。"
  },
  "settings": {
    "storageQuota": "ストレージ容量",
    "storageQuotaHelp": "このユーザーの最大ストレージ容量を設定します。10M、5Gなどの形式を使用してください。0に設定するか空白のままにすると無制限になります（管理者のみ）。",
    "unlimited": "無制限"
  },
  "sidebar": {
    "storageUsed": "{total}中{used}使用",
    "storageUsedUnlimited": "使用中（無制限）"
  }
}
EOF

# Korean
cat > /tmp/ko_patch.json << 'EOF'
{
  "errors": {
    "quotaExceeded": "저장 공간 할당량을 초과했습니다. 일부 파일을 삭제하거나 관리자에게 문의하세요.",
    "quotaExceededDuringUpload": "업로드 중지됨: 저장 공간 할당량 초과. 부분 파일이 삭제되었습니다.",
    "insufficientSpace": "이 업로드에 사용할 수 있는 저장 공간이 부족합니다."
  },
  "settings": {
    "storageQuota": "저장 공간 할당량",
    "storageQuotaHelp": "이 사용자의 최대 저장 공간을 설정합니다. 10M, 5G 등의 형식을 사용하세요. 0으로 설정하거나 비워두면 무제한입니다(관리자만).",
    "unlimited": "무제한"
  },
  "sidebar": {
    "storageUsed": "{total} 중 사용됨",
    "storageUsedUnlimited": "사용됨 (무제한)"
  }
}
EOF

# Portuguese (Brazil)
cat > /tmp/pt-br_patch.json << 'EOF'
{
  "errors": {
    "quotaExceeded": "Cota de armazenamento excedida. Exclua alguns arquivos ou entre em contato com o administrador.",
    "quotaExceededDuringUpload": "Upload interrompido: cota de armazenamento excedida. O arquivo parcial foi excluído.",
    "insufficientSpace": "Espaço de armazenamento insuficiente para este upload."
  },
  "settings": {
    "storageQuota": "Cota de Armazenamento",
    "storageQuotaHelp": "Defina o espaço máximo de armazenamento para este usuário. Use formatos como 10M, 5G, etc. Defina como 0 ou deixe vazio para ilimitado (somente administrador).",
    "unlimited": "Ilimitado"
  },
  "sidebar": {
    "storageUsed": "de {total} usado",
    "storageUsedUnlimited": "usado (Ilimitado)"
  }
}
EOF

# Russian
cat > /tmp/ru_patch.json << 'EOF'
{
  "errors": {
    "quotaExceeded": "Квота хранилища превышена. Удалите некоторые файлы или обратитесь к администратору.",
    "quotaExceededDuringUpload": "Загрузка остановлена: квота хранилища превышена. Частичный файл был удален.",
    "insufficientSpace": "Недостаточно места для этой загрузки."
  },
  "settings": {
    "storageQuota": "Квота Хранилища",
    "storageQuotaHelp": "Установите максимальное пространство хранения для этого пользователя. Используйте форматы типа 10M, 5G и т.д. Установите 0 или оставьте пустым для неограниченного (только администратор).",
    "unlimited": "Неограниченно"
  },
  "sidebar": {
    "storageUsed": "из {total} использовано",
    "storageUsedUnlimited": "использовано (Неограниченно)"
  }
}
EOF

# Chinese Traditional
cat > /tmp/zh-tw_patch.json << 'EOF'
{
  "errors": {
    "quotaExceeded": "儲存配額已超出。請刪除一些檔案或聯絡管理員。",
    "quotaExceededDuringUpload": "上傳已停止：儲存配額已超出。已刪除部分上傳的檔案。",
    "insufficientSpace": "沒有足夠的儲存空間用於此次上傳。"
  },
  "settings": {
    "storageQuota": "儲存配額",
    "storageQuotaHelp": "設定此使用者的最大儲存空間。使用類似 10M、5G 的格式。設定為 0 或留空表示無限制（僅限管理員）。",
    "unlimited": "無限制"
  },
  "sidebar": {
    "storageUsed": "/ {total}",
    "storageUsedUnlimited": "（無限制）"
  }
}
EOF

echo "Applying translations..."

for lang in it ja ko pt-br ru zh-tw; do
  echo "Processing $lang.json..."
  node -e "
    const fs = require('fs');
    const file = '$lang.json';
    const patchFile = '/tmp/${lang}_patch.json';
    
    const data = JSON.parse(fs.readFileSync(file, 'utf-8'));
    const patch = JSON.parse(fs.readFileSync(patchFile, 'utf-8'));
    
    // Merge errors
    Object.assign(data.errors, patch.errors);
    
    // Merge settings
    Object.assign(data.settings, patch.settings);
    
    // Merge sidebar
    Object.assign(data.sidebar, patch.sidebar);
    
    fs.writeFileSync(file, JSON.stringify(data, null, 2) + '\\n');
    console.log('Updated ' + file);
  "
done

echo "All translations completed!"
