/* AllDrop web UI. Written in ES5 on purpose: it must run in IE11 (Windows 7),
   old iOS Safari and Android WebView without any build step or library. */
(function () {
  'use strict';

  var L = {
    tr: {
      devices: 'Yakındaki cihazlar', scan: 'Yeniden tara', scanning: 'Taranıyor…',
      emptyOwner: 'Aynı ağda AllDrop açık başka bir cihaz aranıyor… Telefon ya da tablet bağlamak için aşağıdaki QR kodu okutun.',
      hintOwner: 'Bir cihaza dokunun ya da dosyaları üzerine sürükleyip bırakın.',
      hintGuest: 'Bir cihaza dokunun ve göndermek istediğiniz dosyaları seçin.',
      transfers: 'Aktarımlar', inbox: 'Size gelen dosyalar', zipAll: 'Tümünü ZIP indir', clear: 'Temizle',
      join: 'Telefon veya tablet bağlayın',
      joinSub: 'Uygulaması olmayan cihazlar (telefon, tablet, bilgisayar) hiçbir şey kurmadan tarayıcıdan bağlanır.',
      join1: 'Diğer cihazı bununla aynı Wi-Fi ağına bağlayın.',
      join2: 'Kamerayla QR kodu okutun ya da aşağıdaki adresi tarayıcıya yazın.',
      join3: 'Açılan sayfadan dosya gönderin ve alın.',
      otherNet: 'Diğer ağ adresi', history: 'Son alınanlar', openFolder: 'Klasörü aç',
      pickFiles: 'Dosya ekle', pickDir: 'Klasör ekle', pickMedia: 'Fotoğraf / video',
      trayTitle: 'Gönderilecekler', trayHint: 'Hepsini birlikte göndermek için bir cihaza dokunun. İstediğiniz kadar dosya ve klasör ekleyebilirsiniz.',
      nFiles: '{n} dosya', nFolders: '{n} klasör', more: '… ve {n} öğe daha',
      chooseTitle: '{name} cihazına ne göndermek istiyorsunuz?', chooseHint: 'Birden fazla dosya seçebilir ya da klasörlerin tamamını gönderebilirsiniz. Seçtikleriniz hemen gönderilir.',
      chooseFiles: 'Dosyalar', chooseFolder: 'Klasör', chooseMedia: 'Fotoğraf ve videolar', laterTray: 'Önce listeye ekle',
      zipOffer: 'Hepsini indir (ZIP)', offerHead: '{name} · {n} dosya · {size}', showAll: 'Tümünü göster ({n})', preparing: 'Dosyalar hazırlanıyor…',
      dropHere: 'Bırakın, sonra bir cihaz seçin', dropOn: '{name} cihazına gönder',
      picked: '{n} dosya seçildi · {size}', pickedSub: 'Göndermek için bir cihaza dokunun.',
      staged: '{n} dosya gönderime hazır · {size}',
      offline: 'Bağlantı kesildi. Yeniden bağlanılıyor…',
      lockTitle: 'Bağlantı gerekli',
      lockText: 'Bu sayfayı kullanmak için AllDrop açık olan cihazın ekranındaki QR kodu bu cihazın kamerasıyla okutun.',
      quitTitle: 'AllDrop kapatıldı', quitText: 'Yeniden kullanmak için uygulamayı tekrar açın.',
      offerTitle: '{name} size dosya göndermek istiyor', offerSub: '{n} dosya · {size}',
      accept: 'Kabul et', decline: 'Reddet', trust: 'Bu cihazdan her zaman sormadan kabul et',
      codeShow: 'Gönderimi başlatmak için bu kodu {name} cihazına girin.', codeTrust: 'Bu cihaza güven (bir dahaki sefere kod sorma)',
      codeTitle: '{name} ekranındaki kodu girin', codeHint: 'Alıcının ekranında 6 haneli bir güvenlik kodu göründü. Kod doğru girilince aktarım başlar.',
      codeWrong: 'Kod yanlış. Kalan deneme: {n}', codeSend: 'Gönder', codeEnter: 'Kodu gir', needCode: 'Kod bekleniyor',
      requireCode: 'Gönderenlerden güvenlik kodu iste (önerilen)', requireCodeHint: 'Her gönderimde ekranınızda 6 haneli bir kod çıkar; gönderen bu kodu girmeden aktarım başlamaz. Güvendiğiniz cihazlar kod sormadan gönderebilir.',
      starting: 'Hazırlanıyor', pending: 'Onay bekleniyor', sending: 'Gönderiliyor', accepted: 'Alınıyor',
      done: 'Tamamlandı', declined: 'Reddedildi', expired: 'Zaman aşımı', canceled: 'İptal edildi', error: 'Hata',
      cancel: 'İptal', close: 'Kapat', to: '{name} cihazına', from: '{name} cihazından',
      download: 'İndir', open: 'Aç', remove: 'Sil', show: 'Göster',
      settings: 'Ayarlar', name: 'Cihaz adı', save: 'Kaydet',
      dir: 'Alınan dosyaların klasörü', dirBad: 'Bu klasöre yazılamıyor. Dosyalar geçici klasöre kaydediliyor.',
      autoAccept: 'Gelen dosyaları sormadan kabul et', autoAcceptHint: 'Sadece güvendiğiniz ağlarda açın.',
      trusted: 'Güvenilen cihazlar', none: 'Yok',
      addPeer: 'IP adresiyle cihaz ekle',
      addPeerHint: 'Bir cihaz otomatik görünmüyorsa, o cihazın Ayarlar ekranındaki adresi yazın (ör. 192.168.1.20).',
      add: 'Ekle', adding: 'Aranıyor…',
      firewall: 'Windows Güvenlik Duvarı izni ver', firewallHint: 'Telefonlar bu bilgisayara bağlanamıyorsa kullanın. Yönetici onayı istenir.',
      sendto: 'Sağ tık “Gönder” menüsüne ekle', sendtoHint: 'Gezgin’de dosyaya sağ tıklayıp Gönder → AllDrop ile paylaşın.',
      resetKey: 'QR bağlantısını yenile', resetKeyHint: 'Daha önce bağlanmış tüm tarayıcıların erişimi kapanır.', sure: 'Emin misiniz? Tekrar dokunun',
      quit: 'AllDrop’i kapat', myAddr: 'Bu cihazın adresi', language: 'Dil',
      rename: 'Bu cihazın adı', renameHint: 'Diğer cihazlarda bu ad görünür.',
      saved: 'Kaydedildi', added: 'Cihaz eklendi', notFound: 'O adreste AllDrop bulunamadı',
      receivedN: '{n} dosya alındı', sentOk: '{name} cihazına gönderildi',
      fwOk: 'Açılan pencerede Evet’e tıklayın.', sendtoOk: 'Eklendi: dosyaya sağ tıklayın → Gönder → AllDrop',
      secure: 'Uçtan uca şifreli aktarım (karşılıklı TLS) · Dosyalar yerel ağınızdan dışarı çıkmaz',
      getApps: "Diğer cihazlara AllDrop gönder / indir",
      getAppsTitle: 'Diğer cihazlara AllDrop',
      getAppsSub: 'Bu linkleri herhangi bir cihazda açıp indirin veya paylaşın. Her zaman en güncel sürüm.',
      getAppsShareText: 'AllDrop ile cihazların arasında kablosuz, şifreli dosya gönder:',
      dlPage: 'İndirme sayfası',
      shareLink: 'Paylaş',
      copyLinks: 'Linkleri kopyala',
      copied: 'Kopyalandı',
      version: 'Sürüm', perSec: '/sn', left: '{t} kaldı', justNow: 'az önce', minAgo: '{n} dk önce', hourAgo: '{n} sa önce',
      sec: 'sn', min: 'dk', browserTag: 'tarayıcı', thisHub: 'bu ağın merkezi',
      cantSelf: 'Kendinize gönderemezsiniz', hidden: 'Liste temizlendi',
      os_windows: 'Windows', os_mac: 'Mac', os_linux: 'Linux', os_android: 'Android', os_ios: 'iPhone', os_ipad: 'iPad', os_browser: 'Tarayıcı'
    },
    en: {
      devices: 'Nearby devices', scan: 'Scan again', scanning: 'Scanning…',
      emptyOwner: 'Looking for other devices running AllDrop on this network… To connect a phone or tablet, scan the QR code below.',
      hintOwner: 'Tap a device, or drag and drop files onto it.',
      hintGuest: 'Tap a device and choose the files to send.',
      transfers: 'Transfers', inbox: 'Files for you', zipAll: 'Download all (ZIP)', clear: 'Clear',
      join: 'Connect a phone or tablet',
      joinSub: 'Devices without the app (phones, tablets, computers) connect from the browser, nothing to install.',
      join1: 'Put the other device on the same Wi-Fi network as this one.',
      join2: 'Scan the QR code with the camera, or type the address below into the browser.',
      join3: 'Send and receive files on the page that opens.',
      otherNet: 'Other network address', history: 'Recently received', openFolder: 'Open folder',
      pickFiles: 'Add files', pickDir: 'Add folder', pickMedia: 'Photos / videos',
      trayTitle: 'To send', trayHint: 'Tap a device to send everything at once. Add as many files and folders as you like.',
      nFiles: '{n} files', nFolders: '{n} folders', nFiles1: '1 file', nFolders1: '1 folder', more: '… and {n} more',
      chooseTitle: 'What do you want to send to {name}?', chooseHint: 'Pick several files or send whole folders. Your selection is sent right away.',
      chooseFiles: 'Files', chooseFolder: 'Folder', chooseMedia: 'Photos and videos', laterTray: 'Add to the list first',
      zipOffer: 'Download all (ZIP)', offerHead: '{name} · {n} files · {size}', showAll: 'Show all ({n})', preparing: 'Preparing files…',
      dropHere: 'Drop, then pick a device', dropOn: 'Send to {name}',
      picked: '{n} files selected · {size}', pickedSub: 'Tap a device to send them.',
      staged: '{n} files ready to send · {size}',
      offline: 'Connection lost. Reconnecting…',
      lockTitle: 'Connection required',
      lockText: 'To use this page, scan the QR code shown by AllDrop on the other device with this device’s camera.',
      quitTitle: 'AllDrop has been closed', quitText: 'Open the app again to keep using it.',
      offerTitle: '{name} wants to send you files', offerSub: '{n} files · {size}',
      accept: 'Accept', decline: 'Decline', trust: 'Always accept from this device without asking',
      codeShow: 'Enter this code on {name} to start the transfer.', codeTrust: 'Trust this device (no code next time)',
      codeTitle: 'Enter the code shown on {name}', codeHint: 'A 6-digit security code appeared on the receiver’s screen. The transfer starts once the code is entered correctly.',
      codeWrong: 'Wrong code. Tries left: {n}', codeSend: 'Send', codeEnter: 'Enter code', needCode: 'Waiting for code',
      requireCode: 'Ask senders for a security code (recommended)', requireCodeHint: 'For each transfer a 6-digit code appears on your screen; nothing is sent until the sender types it. Trusted devices can send without a code.',
      starting: 'Preparing', pending: 'Waiting for approval', sending: 'Sending', accepted: 'Receiving',
      done: 'Done', declined: 'Declined', expired: 'Timed out', canceled: 'Canceled', error: 'Error',
      cancel: 'Cancel', close: 'Close', to: 'to {name}', from: 'from {name}',
      download: 'Download', open: 'Open', remove: 'Delete', show: 'Show',
      settings: 'Settings', name: 'Device name', save: 'Save',
      dir: 'Folder for received files', dirBad: 'This folder is not writable. Files are saved to a temporary folder.',
      autoAccept: 'Accept incoming files without asking', autoAcceptHint: 'Only turn this on in networks you trust.',
      trusted: 'Trusted devices', none: 'None',
      addPeer: 'Add a device by IP address',
      addPeerHint: 'If a device does not show up automatically, type the address shown in its Settings (e.g. 192.168.1.20).',
      add: 'Add', adding: 'Looking…',
      firewall: 'Allow through Windows Firewall', firewallHint: 'Use this if phones cannot connect to this PC. Asks for administrator approval.',
      sendto: 'Add to the right-click “Send to” menu', sendtoHint: 'Right-click a file in Explorer → Send to → AllDrop.',
      resetKey: 'Renew QR link', resetKeyHint: 'All browsers that joined before lose access.', sure: 'Sure? Tap again',
      quit: 'Quit AllDrop', myAddr: 'This device’s address', language: 'Language',
      rename: 'Name of this device', renameHint: 'Other devices see this name.',
      saved: 'Saved', added: 'Device added', notFound: 'No AllDrop answered at that address',
      receivedN: '{n} files received', sentOk: 'Sent to {name}',
      fwOk: 'Click Yes in the window that opens.', sendtoOk: 'Added: right-click a file → Send to → AllDrop',
      secure: 'End-to-end encrypted transfer (mutual TLS) · Files never leave your local network',
      getApps: 'Get / share AllDrop on your other devices',
      getAppsTitle: 'Get AllDrop on your other devices',
      getAppsSub: 'Open these links on any device to download, or share them. Always the latest version.',
      getAppsShareText: 'Send files wirelessly and encrypted between your devices with AllDrop:',
      dlPage: 'Download page',
      shareLink: 'Share',
      copyLinks: 'Copy links',
      copied: 'Copied',
      version: 'Version', perSec: '/s', left: '{t} left', justNow: 'just now', minAgo: '{n} min ago', hourAgo: '{n} h ago',
      sec: 's', min: 'min', browserTag: 'browser', thisHub: 'network hub',
      cantSelf: 'You cannot send to yourself', hidden: 'List cleared',
      os_windows: 'Windows', os_mac: 'Mac', os_linux: 'Linux', os_android: 'Android', os_ios: 'iPhone', os_ipad: 'iPad', os_browser: 'Browser'
    },
    es: {"devices": "Dispositivos cercanos", "scan": "Buscar de nuevo", "scanning": "Buscando…", "emptyOwner": "Buscando otros dispositivos con AllDrop en esta red… Para conectar un teléfono o tableta, escanea el código QR de abajo.", "hintOwner": "Toca un dispositivo o arrastra y suelta archivos sobre él.", "hintGuest": "Toca un dispositivo y elige los archivos que enviar.", "transfers": "Transferencias", "inbox": "Archivos para ti", "zipAll": "Descargar todo (ZIP)", "clear": "Borrar", "join": "Conectar un teléfono o tableta", "joinSub": "Los dispositivos sin la app (teléfonos, tabletas, ordenadores) se conectan desde el navegador, sin instalar nada.", "join1": "Pon el otro dispositivo en la misma red Wi-Fi que este.", "join2": "Escanea el código QR con la cámara o escribe la dirección de abajo en el navegador.", "join3": "Envía y recibe archivos en la página que se abre.", "otherNet": "Otra dirección de red", "history": "Recibidos recientemente", "openFolder": "Abrir carpeta", "pickFiles": "Añadir archivos", "pickDir": "Añadir carpeta", "pickMedia": "Fotos / vídeos", "trayTitle": "Para enviar", "trayHint": "Toca un dispositivo para enviarlo todo a la vez. Añade tantos archivos y carpetas como quieras.", "nFiles": "{n} archivos", "nFolders": "{n} carpetas", "nFiles1": "1 archivo", "nFolders1": "1 carpeta", "more": "… y {n} más", "chooseTitle": "¿Qué quieres enviar a {name}?", "chooseHint": "Elige varios archivos o envía carpetas completas. Lo que elijas se envía enseguida.", "chooseFiles": "Archivos", "chooseFolder": "Carpeta", "chooseMedia": "Fotos y vídeos", "laterTray": "Añadir primero a la lista", "zipOffer": "Descargar todo (ZIP)", "offerHead": "{name} · {n} archivos · {size}", "showAll": "Mostrar todo ({n})", "preparing": "Preparando archivos…", "dropHere": "Suelta y elige un dispositivo", "dropOn": "Enviar a {name}", "picked": "{n} archivos seleccionados · {size}", "pickedSub": "Toca un dispositivo para enviarlos.", "staged": "{n} archivos listos para enviar · {size}", "offline": "Conexión perdida. Reconectando…", "lockTitle": "Se requiere conexión", "lockText": "Para usar esta página, escanea con la cámara de este dispositivo el código QR que muestra AllDrop en el otro dispositivo.", "quitTitle": "AllDrop se ha cerrado", "quitText": "Abre la app de nuevo para seguir usándola.", "offerTitle": "{name} quiere enviarte archivos", "offerSub": "{n} archivos · {size}", "accept": "Aceptar", "decline": "Rechazar", "trust": "Aceptar siempre de este dispositivo sin preguntar", "codeShow": "Introduce este código en {name} para iniciar la transferencia.", "codeTrust": "Confiar en este dispositivo (sin código la próxima vez)", "codeTitle": "Introduce el código que aparece en {name}", "codeHint": "Apareció un código de seguridad de 6 dígitos en la pantalla del receptor. La transferencia empieza cuando se introduce correctamente.", "codeWrong": "Código incorrecto. Intentos restantes: {n}", "codeSend": "Enviar", "codeEnter": "Introducir código", "needCode": "Esperando el código", "requireCode": "Pedir un código de seguridad a quien envía (recomendado)", "requireCodeHint": "En cada transferencia aparece un código de 6 dígitos en tu pantalla; no se envía nada hasta que quien envía lo escribe. Los dispositivos de confianza pueden enviar sin código.", "starting": "Preparando", "pending": "Esperando aprobación", "sending": "Enviando", "accepted": "Recibiendo", "done": "Hecho", "declined": "Rechazado", "expired": "Tiempo agotado", "canceled": "Cancelado", "error": "Error", "cancel": "Cancelar", "close": "Cerrar", "to": "a {name}", "from": "de {name}", "download": "Descargar", "open": "Abrir", "remove": "Eliminar", "show": "Mostrar", "settings": "Ajustes", "name": "Nombre del dispositivo", "save": "Guardar", "dir": "Carpeta para archivos recibidos", "dirBad": "No se puede escribir en esta carpeta. Los archivos se guardan en una carpeta temporal.", "autoAccept": "Aceptar archivos entrantes sin preguntar", "autoAcceptHint": "Actívalo solo en redes de confianza.", "trusted": "Dispositivos de confianza", "none": "Ninguno", "addPeer": "Añadir un dispositivo por dirección IP", "addPeerHint": "Si un dispositivo no aparece automáticamente, escribe la dirección que se muestra en sus Ajustes (p. ej. 192.168.1.20).", "add": "Añadir", "adding": "Buscando…", "firewall": "Permitir en el Firewall de Windows", "firewallHint": "Úsalo si los teléfonos no pueden conectarse a este PC. Pide aprobación de administrador.", "sendto": "Añadir al menú “Enviar a” del clic derecho", "sendtoHint": "Haz clic derecho en un archivo en el Explorador → Enviar a → AllDrop.", "resetKey": "Renovar enlace QR", "resetKeyHint": "Todos los navegadores que se unieron antes pierden el acceso.", "sure": "¿Seguro? Toca otra vez", "quit": "Salir de AllDrop", "myAddr": "Dirección de este dispositivo", "language": "Idioma", "rename": "Nombre de este dispositivo", "renameHint": "Otros dispositivos ven este nombre.", "saved": "Guardado", "added": "Dispositivo añadido", "notFound": "Ningún AllDrop respondió en esa dirección", "receivedN": "{n} archivos recibidos", "sentOk": "Enviado a {name}", "fwOk": "Haz clic en Sí en la ventana que se abre.", "sendtoOk": "Añadido: clic derecho en un archivo → Enviar a → AllDrop", "secure": "Transferencia cifrada de extremo a extremo (TLS mutuo) · Los archivos no salen de tu red local", "getApps": "Consigue AllDrop para tus otros dispositivos", "version": "Versión", "perSec": "/s", "left": "{t} restante", "justNow": "ahora mismo", "minAgo": "hace {n} min", "hourAgo": "hace {n} h", "sec": "s", "min": "min", "browserTag": "navegador", "thisHub": "nodo de red", "cantSelf": "No puedes enviarte a ti mismo", "hidden": "Lista borrada", "os_browser": "Navegador", "os_windows": "Windows", "os_mac": "Mac", "os_linux": "Linux", "os_android": "Android", "os_ios": "iPhone", "os_ipad": "iPad"},
    de: {"devices": "Geräte in der Nähe", "scan": "Erneut suchen", "scanning": "Suche läuft…", "emptyOwner": "Suche nach anderen Geräten mit AllDrop in diesem Netzwerk… Um ein Smartphone oder Tablet zu verbinden, scanne den QR-Code unten.", "hintOwner": "Tippe auf ein Gerät oder ziehe Dateien darauf.", "hintGuest": "Tippe auf ein Gerät und wähle die zu sendenden Dateien.", "transfers": "Übertragungen", "inbox": "Dateien für dich", "zipAll": "Alle herunterladen (ZIP)", "clear": "Leeren", "join": "Smartphone oder Tablet verbinden", "joinSub": "Geräte ohne die App (Smartphones, Tablets, Computer) verbinden sich über den Browser, ohne Installation.", "join1": "Bring das andere Gerät ins selbe WLAN wie dieses.", "join2": "Scanne den QR-Code mit der Kamera oder gib die Adresse unten im Browser ein.", "join3": "Sende und empfange Dateien auf der Seite, die sich öffnet.", "otherNet": "Andere Netzwerkadresse", "history": "Zuletzt empfangen", "openFolder": "Ordner öffnen", "pickFiles": "Dateien hinzufügen", "pickDir": "Ordner hinzufügen", "pickMedia": "Fotos / Videos", "trayTitle": "Zu senden", "trayHint": "Tippe auf ein Gerät, um alles auf einmal zu senden. Füge beliebig viele Dateien und Ordner hinzu.", "nFiles": "{n} Dateien", "nFolders": "{n} Ordner", "nFiles1": "1 Datei", "nFolders1": "1 Ordner", "more": "… und {n} weitere", "chooseTitle": "Was möchtest du an {name} senden?", "chooseHint": "Wähle mehrere Dateien oder sende ganze Ordner. Deine Auswahl wird sofort gesendet.", "chooseFiles": "Dateien", "chooseFolder": "Ordner", "chooseMedia": "Fotos und Videos", "laterTray": "Erst zur Liste hinzufügen", "zipOffer": "Alle herunterladen (ZIP)", "offerHead": "{name} · {n} Dateien · {size}", "showAll": "Alle anzeigen ({n})", "preparing": "Dateien werden vorbereitet…", "dropHere": "Ablegen, dann Gerät wählen", "dropOn": "An {name} senden", "picked": "{n} Dateien ausgewählt · {size}", "pickedSub": "Tippe auf ein Gerät, um sie zu senden.", "staged": "{n} Dateien bereit zum Senden · {size}", "offline": "Verbindung verloren. Neu verbinden…", "lockTitle": "Verbindung erforderlich", "lockText": "Um diese Seite zu nutzen, scanne mit der Kamera dieses Geräts den QR-Code, den AllDrop auf dem anderen Gerät zeigt.", "quitTitle": "AllDrop wurde geschlossen", "quitText": "Öffne die App erneut, um sie weiter zu nutzen.", "offerTitle": "{name} möchte dir Dateien senden", "offerSub": "{n} Dateien · {size}", "accept": "Annehmen", "decline": "Ablehnen", "trust": "Von diesem Gerät immer ohne Nachfrage annehmen", "codeShow": "Gib diesen Code auf {name} ein, um die Übertragung zu starten.", "codeTrust": "Diesem Gerät vertrauen (nächstes Mal kein Code)", "codeTitle": "Gib den auf {name} angezeigten Code ein", "codeHint": "Auf dem Bildschirm des Empfängers erschien ein 6-stelliger Sicherheitscode. Die Übertragung startet, sobald er korrekt eingegeben wird.", "codeWrong": "Falscher Code. Verbleibende Versuche: {n}", "codeSend": "Senden", "codeEnter": "Code eingeben", "needCode": "Warte auf Code", "requireCode": "Sender nach einem Sicherheitscode fragen (empfohlen)", "requireCodeHint": "Bei jeder Übertragung erscheint ein 6-stelliger Code auf deinem Bildschirm; es wird nichts gesendet, bis der Sender ihn eingibt. Vertrauenswürdige Geräte können ohne Code senden.", "starting": "Vorbereiten", "pending": "Warte auf Bestätigung", "sending": "Senden", "accepted": "Empfangen", "done": "Fertig", "declined": "Abgelehnt", "expired": "Zeitüberschreitung", "canceled": "Abgebrochen", "error": "Fehler", "cancel": "Abbrechen", "close": "Schließen", "to": "an {name}", "from": "von {name}", "download": "Herunterladen", "open": "Öffnen", "remove": "Löschen", "show": "Anzeigen", "settings": "Einstellungen", "name": "Gerätename", "save": "Speichern", "dir": "Ordner für empfangene Dateien", "dirBad": "In diesen Ordner kann nicht geschrieben werden. Dateien werden in einem temporären Ordner gespeichert.", "autoAccept": "Eingehende Dateien ohne Nachfrage annehmen", "autoAcceptHint": "Aktiviere dies nur in Netzwerken, denen du vertraust.", "trusted": "Vertrauenswürdige Geräte", "none": "Keine", "addPeer": "Gerät per IP-Adresse hinzufügen", "addPeerHint": "Wenn ein Gerät nicht automatisch erscheint, gib die in seinen Einstellungen angezeigte Adresse ein (z. B. 192.168.1.20).", "add": "Hinzufügen", "adding": "Suche…", "firewall": "Durch die Windows-Firewall erlauben", "firewallHint": "Nutze dies, wenn Smartphones sich nicht mit diesem PC verbinden können. Fragt nach Administratorbestätigung.", "sendto": "Zum Rechtsklick-Menü „Senden an“ hinzufügen", "sendtoHint": "Rechtsklick auf eine Datei im Explorer → Senden an → AllDrop.", "resetKey": "QR-Link erneuern", "resetKeyHint": "Alle zuvor verbundenen Browser verlieren den Zugriff.", "sure": "Sicher? Nochmal tippen", "quit": "AllDrop beenden", "myAddr": "Adresse dieses Geräts", "language": "Sprache", "rename": "Name dieses Geräts", "renameHint": "Andere Geräte sehen diesen Namen.", "saved": "Gespeichert", "added": "Gerät hinzugefügt", "notFound": "Unter dieser Adresse hat kein AllDrop geantwortet", "receivedN": "{n} Dateien empfangen", "sentOk": "An {name} gesendet", "fwOk": "Klicke im Fenster, das sich öffnet, auf Ja.", "sendtoOk": "Hinzugefügt: Rechtsklick auf eine Datei → Senden an → AllDrop", "secure": "Ende-zu-Ende-verschlüsselte Übertragung (gegenseitiges TLS) · Dateien verlassen dein lokales Netzwerk nie", "getApps": "AllDrop für deine anderen Geräte holen", "version": "Version", "perSec": "/s", "left": "noch {t}", "justNow": "gerade eben", "minAgo": "vor {n} Min", "hourAgo": "vor {n} Std", "sec": "s", "min": "Min", "browserTag": "Browser", "thisHub": "Netzwerk-Hub", "cantSelf": "Du kannst nicht an dich selbst senden", "hidden": "Liste geleert", "os_browser": "Browser", "os_windows": "Windows", "os_mac": "Mac", "os_linux": "Linux", "os_android": "Android", "os_ios": "iPhone", "os_ipad": "iPad"},
    fr: {"devices": "Appareils à proximité", "scan": "Rechercher à nouveau", "scanning": "Recherche…", "emptyOwner": "Recherche d’autres appareils exécutant AllDrop sur ce réseau… Pour connecter un téléphone ou une tablette, scannez le code QR ci-dessous.", "hintOwner": "Touchez un appareil, ou glissez-déposez des fichiers dessus.", "hintGuest": "Touchez un appareil et choisissez les fichiers à envoyer.", "transfers": "Transferts", "inbox": "Fichiers pour vous", "zipAll": "Tout télécharger (ZIP)", "clear": "Effacer", "join": "Connecter un téléphone ou une tablette", "joinSub": "Les appareils sans l’app (téléphones, tablettes, ordinateurs) se connectent depuis le navigateur, rien à installer.", "join1": "Mettez l’autre appareil sur le même réseau Wi-Fi que celui-ci.", "join2": "Scannez le code QR avec la caméra, ou saisissez l’adresse ci-dessous dans le navigateur.", "join3": "Envoyez et recevez des fichiers sur la page qui s’ouvre.", "otherNet": "Autre adresse réseau", "history": "Reçus récemment", "openFolder": "Ouvrir le dossier", "pickFiles": "Ajouter des fichiers", "pickDir": "Ajouter un dossier", "pickMedia": "Photos / vidéos", "trayTitle": "À envoyer", "trayHint": "Touchez un appareil pour tout envoyer d’un coup. Ajoutez autant de fichiers et de dossiers que vous voulez.", "nFiles": "{n} fichiers", "nFolders": "{n} dossiers", "nFiles1": "1 fichier", "nFolders1": "1 dossier", "more": "… et {n} de plus", "chooseTitle": "Que voulez-vous envoyer à {name} ?", "chooseHint": "Choisissez plusieurs fichiers ou envoyez des dossiers entiers. Votre sélection part tout de suite.", "chooseFiles": "Fichiers", "chooseFolder": "Dossier", "chooseMedia": "Photos et vidéos", "laterTray": "D’abord ajouter à la liste", "zipOffer": "Tout télécharger (ZIP)", "offerHead": "{name} · {n} fichiers · {size}", "showAll": "Tout afficher ({n})", "preparing": "Préparation des fichiers…", "dropHere": "Déposez, puis choisissez un appareil", "dropOn": "Envoyer à {name}", "picked": "{n} fichiers sélectionnés · {size}", "pickedSub": "Touchez un appareil pour les envoyer.", "staged": "{n} fichiers prêts à envoyer · {size}", "offline": "Connexion perdue. Reconnexion…", "lockTitle": "Connexion requise", "lockText": "Pour utiliser cette page, scannez avec la caméra de cet appareil le code QR affiché par AllDrop sur l’autre appareil.", "quitTitle": "AllDrop a été fermé", "quitText": "Rouvrez l’app pour continuer à l’utiliser.", "offerTitle": "{name} veut vous envoyer des fichiers", "offerSub": "{n} fichiers · {size}", "accept": "Accepter", "decline": "Refuser", "trust": "Toujours accepter de cet appareil sans demander", "codeShow": "Saisissez ce code sur {name} pour démarrer le transfert.", "codeTrust": "Faire confiance à cet appareil (pas de code la prochaine fois)", "codeTitle": "Saisissez le code affiché sur {name}", "codeHint": "Un code de sécurité à 6 chiffres est apparu sur l’écran du destinataire. Le transfert démarre une fois le code saisi correctement.", "codeWrong": "Code incorrect. Essais restants : {n}", "codeSend": "Envoyer", "codeEnter": "Saisir le code", "needCode": "En attente du code", "requireCode": "Demander un code de sécurité aux expéditeurs (recommandé)", "requireCodeHint": "À chaque transfert, un code à 6 chiffres apparaît sur votre écran ; rien n’est envoyé tant que l’expéditeur ne l’a pas saisi. Les appareils de confiance peuvent envoyer sans code.", "starting": "Préparation", "pending": "En attente d’approbation", "sending": "Envoi", "accepted": "Réception", "done": "Terminé", "declined": "Refusé", "expired": "Expiré", "canceled": "Annulé", "error": "Erreur", "cancel": "Annuler", "close": "Fermer", "to": "à {name}", "from": "de {name}", "download": "Télécharger", "open": "Ouvrir", "remove": "Supprimer", "show": "Afficher", "settings": "Réglages", "name": "Nom de l’appareil", "save": "Enregistrer", "dir": "Dossier des fichiers reçus", "dirBad": "Ce dossier n’est pas accessible en écriture. Les fichiers sont enregistrés dans un dossier temporaire.", "autoAccept": "Accepter les fichiers entrants sans demander", "autoAcceptHint": "À n’activer que sur les réseaux de confiance.", "trusted": "Appareils de confiance", "none": "Aucun", "addPeer": "Ajouter un appareil par adresse IP", "addPeerHint": "Si un appareil n’apparaît pas automatiquement, saisissez l’adresse affichée dans ses Réglages (ex. 192.168.1.20).", "add": "Ajouter", "adding": "Recherche…", "firewall": "Autoriser dans le pare-feu Windows", "firewallHint": "Utilisez ceci si les téléphones ne peuvent pas se connecter à ce PC. Demande une approbation administrateur.", "sendto": "Ajouter au menu clic droit « Envoyer vers »", "sendtoHint": "Clic droit sur un fichier dans l’Explorateur → Envoyer vers → AllDrop.", "resetKey": "Renouveler le lien QR", "resetKeyHint": "Tous les navigateurs déjà connectés perdent l’accès.", "sure": "Sûr ? Touchez à nouveau", "quit": "Quitter AllDrop", "myAddr": "Adresse de cet appareil", "language": "Langue", "rename": "Nom de cet appareil", "renameHint": "Les autres appareils voient ce nom.", "saved": "Enregistré", "added": "Appareil ajouté", "notFound": "Aucun AllDrop n’a répondu à cette adresse", "receivedN": "{n} fichiers reçus", "sentOk": "Envoyé à {name}", "fwOk": "Cliquez sur Oui dans la fenêtre qui s’ouvre.", "sendtoOk": "Ajouté : clic droit sur un fichier → Envoyer vers → AllDrop", "secure": "Transfert chiffré de bout en bout (TLS mutuel) · Les fichiers ne quittent jamais votre réseau local", "getApps": "Obtenir AllDrop pour vos autres appareils", "version": "Version", "perSec": "/s", "left": "{t} restant", "justNow": "à l’instant", "minAgo": "il y a {n} min", "hourAgo": "il y a {n} h", "sec": "s", "min": "min", "browserTag": "navigateur", "thisHub": "nœud réseau", "cantSelf": "Vous ne pouvez pas vous envoyer à vous-même", "hidden": "Liste effacée", "os_browser": "Navigateur", "os_windows": "Windows", "os_mac": "Mac", "os_linux": "Linux", "os_android": "Android", "os_ios": "iPhone", "os_ipad": "iPad"},
    ar: {"devices": "الأجهزة القريبة", "scan": "إعادة البحث", "scanning": "جارٍ البحث…", "emptyOwner": "جارٍ البحث عن أجهزة أخرى تعمل بـ AllDrop على هذه الشبكة… لتوصيل هاتف أو جهاز لوحي، امسح رمز QR أدناه.", "hintOwner": "انقر على جهاز، أو اسحب الملفات وأفلتها عليه.", "hintGuest": "انقر على جهاز واختر الملفات المراد إرسالها.", "transfers": "عمليات النقل", "inbox": "ملفات لك", "zipAll": "تنزيل الكل (ZIP)", "clear": "مسح", "join": "توصيل هاتف أو جهاز لوحي", "joinSub": "الأجهزة التي لا تحتوي على التطبيق (هواتف، أجهزة لوحية، حواسيب) تتصل من المتصفح دون تثبيت أي شيء.", "join1": "ضع الجهاز الآخر على نفس شبكة Wi-Fi مثل هذا الجهاز.", "join2": "امسح رمز QR بالكاميرا، أو اكتب العنوان أدناه في المتصفح.", "join3": "أرسل واستقبل الملفات في الصفحة التي تفتح.", "otherNet": "عنوان شبكة آخر", "history": "المستلمة مؤخرًا", "openFolder": "فتح المجلد", "pickFiles": "إضافة ملفات", "pickDir": "إضافة مجلد", "pickMedia": "صور / فيديو", "trayTitle": "للإرسال", "trayHint": "انقر على جهاز لإرسال كل شيء دفعة واحدة. أضف ما تشاء من الملفات والمجلدات.", "nFiles": "{n} ملفات", "nFolders": "{n} مجلدات", "nFiles1": "ملف واحد", "nFolders1": "مجلد واحد", "more": "… و{n} أخرى", "chooseTitle": "ماذا تريد أن ترسل إلى {name}؟", "chooseHint": "اختر عدة ملفات أو أرسل مجلدات كاملة. يُرسَل ما تختاره فورًا.", "chooseFiles": "ملفات", "chooseFolder": "مجلد", "chooseMedia": "صور وفيديوهات", "laterTray": "أضف إلى القائمة أولًا", "zipOffer": "تنزيل الكل (ZIP)", "offerHead": "{name} · {n} ملفات · {size}", "showAll": "عرض الكل ({n})", "preparing": "جارٍ تجهيز الملفات…", "dropHere": "أفلت ثم اختر جهازًا", "dropOn": "إرسال إلى {name}", "picked": "تم اختيار {n} ملفات · {size}", "pickedSub": "انقر على جهاز لإرسالها.", "staged": "{n} ملفات جاهزة للإرسال · {size}", "offline": "انقطع الاتصال. جارٍ إعادة الاتصال…", "lockTitle": "الاتصال مطلوب", "lockText": "لاستخدام هذه الصفحة، امسح بكاميرا هذا الجهاز رمز QR الذي يعرضه AllDrop على الجهاز الآخر.", "quitTitle": "تم إغلاق AllDrop", "quitText": "افتح التطبيق مرة أخرى لمواصلة استخدامه.", "offerTitle": "يريد {name} إرسال ملفات إليك", "offerSub": "{n} ملفات · {size}", "accept": "قبول", "decline": "رفض", "trust": "القبول دائمًا من هذا الجهاز دون سؤال", "codeShow": "أدخل هذا الرمز على {name} لبدء النقل.", "codeTrust": "الوثوق بهذا الجهاز (بدون رمز في المرة القادمة)", "codeTitle": "أدخل الرمز المعروض على {name}", "codeHint": "ظهر رمز أمان من 6 أرقام على شاشة المستلم. يبدأ النقل بمجرد إدخال الرمز بشكل صحيح.", "codeWrong": "رمز خاطئ. المحاولات المتبقية: {n}", "codeSend": "إرسال", "codeEnter": "أدخل الرمز", "needCode": "في انتظار الرمز", "requireCode": "طلب رمز أمان من المُرسِلين (موصى به)", "requireCodeHint": "في كل عملية نقل يظهر رمز من 6 أرقام على شاشتك؛ لا يُرسَل شيء حتى يُدخله المُرسِل. يمكن للأجهزة الموثوقة الإرسال دون رمز.", "starting": "جارٍ التحضير", "pending": "في انتظار الموافقة", "sending": "جارٍ الإرسال", "accepted": "جارٍ الاستلام", "done": "تم", "declined": "مرفوض", "expired": "انتهت المهلة", "canceled": "أُلغي", "error": "خطأ", "cancel": "إلغاء", "close": "إغلاق", "to": "إلى {name}", "from": "من {name}", "download": "تنزيل", "open": "فتح", "remove": "حذف", "show": "عرض", "settings": "الإعدادات", "name": "اسم الجهاز", "save": "حفظ", "dir": "مجلد الملفات المستلمة", "dirBad": "لا يمكن الكتابة في هذا المجلد. تُحفظ الملفات في مجلد مؤقت.", "autoAccept": "قبول الملفات الواردة دون سؤال", "autoAcceptHint": "فعّل هذا فقط في الشبكات التي تثق بها.", "trusted": "الأجهزة الموثوقة", "none": "لا شيء", "addPeer": "إضافة جهاز عبر عنوان IP", "addPeerHint": "إذا لم يظهر جهاز تلقائيًا، اكتب العنوان المعروض في إعداداته (مثال: 192.168.1.20).", "add": "إضافة", "adding": "جارٍ البحث…", "firewall": "السماح عبر جدار حماية Windows", "firewallHint": "استخدم هذا إذا تعذّر على الهواتف الاتصال بهذا الحاسوب. يطلب موافقة المسؤول.", "sendto": "إضافة إلى قائمة «إرسال إلى» بالنقر الأيمن", "sendtoHint": "انقر بزر الفأرة الأيمن على ملف في المستكشف ← إرسال إلى ← AllDrop.", "resetKey": "تجديد رابط QR", "resetKeyHint": "تفقد كل المتصفحات التي اتصلت سابقًا صلاحية الوصول.", "sure": "متأكد؟ انقر مرة أخرى", "quit": "إنهاء AllDrop", "myAddr": "عنوان هذا الجهاز", "language": "اللغة", "rename": "اسم هذا الجهاز", "renameHint": "تظهر هذه التسمية على الأجهزة الأخرى.", "saved": "تم الحفظ", "added": "تمت إضافة الجهاز", "notFound": "لم يستجب أي AllDrop على ذلك العنوان", "receivedN": "تم استلام {n} ملفات", "sentOk": "تم الإرسال إلى {name}", "fwOk": "انقر على نعم في النافذة التي تفتح.", "sendtoOk": "تمت الإضافة: انقر بزر الفأرة الأيمن على ملف ← إرسال إلى ← AllDrop", "secure": "نقل مشفّر من طرف إلى طرف (TLS متبادل) · الملفات لا تغادر شبكتك المحلية", "getApps": "احصل على AllDrop لأجهزتك الأخرى", "version": "الإصدار", "perSec": "/ث", "left": "يتبقى {t}", "justNow": "الآن", "minAgo": "قبل {n} دقيقة", "hourAgo": "قبل {n} ساعة", "sec": "ث", "min": "دقيقة", "browserTag": "المتصفح", "thisHub": "موزّع الشبكة", "cantSelf": "لا يمكنك الإرسال إلى نفسك", "hidden": "تم مسح القائمة", "os_browser": "المتصفح", "os_windows": "Windows", "os_mac": "Mac", "os_linux": "Linux", "os_android": "Android", "os_ios": "iPhone", "os_ipad": "iPad"},
    zh: {"devices": "附近的设备", "scan": "重新扫描", "scanning": "正在扫描…", "emptyOwner": "正在此网络中查找运行 AllDrop 的其他设备… 要连接手机或平板，请扫描下方的二维码。", "hintOwner": "点按某个设备，或将文件拖放到其上。", "hintGuest": "点按某个设备并选择要发送的文件。", "transfers": "传输", "inbox": "发给你的文件", "zipAll": "全部下载（ZIP）", "clear": "清除", "join": "连接手机或平板", "joinSub": "没有安装应用的设备（手机、平板、电脑）可通过浏览器连接，无需安装。", "join1": "让另一台设备连接到与此设备相同的 Wi-Fi 网络。", "join2": "用相机扫描二维码，或在浏览器中输入下方地址。", "join3": "在打开的页面上收发文件。", "otherNet": "其他网络地址", "history": "最近接收", "openFolder": "打开文件夹", "pickFiles": "添加文件", "pickDir": "添加文件夹", "pickMedia": "照片 / 视频", "trayTitle": "待发送", "trayHint": "点按一个设备即可一次全部发送。可以添加任意数量的文件和文件夹。", "nFiles": "{n} 个文件", "nFolders": "{n} 个文件夹", "nFiles1": "1 个文件", "nFolders1": "1 个文件夹", "more": "… 还有 {n} 项", "chooseTitle": "要向 {name} 发送什么？", "chooseHint": "可选择多个文件或发送整个文件夹。所选内容会立即发送。", "chooseFiles": "文件", "chooseFolder": "文件夹", "chooseMedia": "照片和视频", "laterTray": "先加入列表", "zipOffer": "全部下载（ZIP）", "offerHead": "{name} · {n} 个文件 · {size}", "showAll": "显示全部（{n}）", "preparing": "正在准备文件…", "dropHere": "拖放后选择设备", "dropOn": "发送到 {name}", "picked": "已选择 {n} 个文件 · {size}", "pickedSub": "点按设备以发送。", "staged": "{n} 个文件已准备发送 · {size}", "offline": "连接已断开。正在重新连接…", "lockTitle": "需要连接", "lockText": "要使用此页面，请用此设备的相机扫描另一台设备上 AllDrop 显示的二维码。", "quitTitle": "AllDrop 已关闭", "quitText": "重新打开应用以继续使用。", "offerTitle": "{name} 想向你发送文件", "offerSub": "{n} 个文件 · {size}", "accept": "接受", "decline": "拒绝", "trust": "始终接受此设备且不再询问", "codeShow": "在 {name} 上输入此代码以开始传输。", "codeTrust": "信任此设备（下次免代码）", "codeTitle": "输入 {name} 上显示的代码", "codeHint": "接收方屏幕上出现了一个 6 位安全代码。正确输入后传输即开始。", "codeWrong": "代码错误。剩余尝试：{n}", "codeSend": "发送", "codeEnter": "输入代码", "needCode": "等待代码", "requireCode": "要求发送方提供安全代码（推荐）", "requireCodeHint": "每次传输都会在你的屏幕上显示一个 6 位代码；在发送方输入之前不会发送任何内容。受信任的设备可免代码发送。", "starting": "准备中", "pending": "等待批准", "sending": "发送中", "accepted": "接收中", "done": "完成", "declined": "已拒绝", "expired": "已超时", "canceled": "已取消", "error": "错误", "cancel": "取消", "close": "关闭", "to": "到 {name}", "from": "来自 {name}", "download": "下载", "open": "打开", "remove": "删除", "show": "显示", "settings": "设置", "name": "设备名称", "save": "保存", "dir": "接收文件的文件夹", "dirBad": "此文件夹不可写入。文件将保存到临时文件夹。", "autoAccept": "无需询问即接受传入文件", "autoAcceptHint": "仅在你信任的网络中开启。", "trusted": "受信任的设备", "none": "无", "addPeer": "通过 IP 地址添加设备", "addPeerHint": "如果设备未自动出现，请输入其“设置”中显示的地址（例如 192.168.1.20）。", "add": "添加", "adding": "查找中…", "firewall": "允许通过 Windows 防火墙", "firewallHint": "如果手机无法连接到此电脑，请使用此项。需要管理员批准。", "sendto": "添加到右键“发送到”菜单", "sendtoHint": "在资源管理器中右键点击文件 → 发送到 → AllDrop。", "resetKey": "更新二维码链接", "resetKeyHint": "所有先前已连接的浏览器将失去访问权限。", "sure": "确定？再次点按", "quit": "退出 AllDrop", "myAddr": "此设备的地址", "language": "语言", "rename": "此设备的名称", "renameHint": "其他设备会看到此名称。", "saved": "已保存", "added": "设备已添加", "notFound": "该地址没有 AllDrop 响应", "receivedN": "已接收 {n} 个文件", "sentOk": "已发送到 {name}", "fwOk": "在打开的窗口中点击“是”。", "sendtoOk": "已添加：右键点击文件 → 发送到 → AllDrop", "secure": "端到端加密传输（双向 TLS）· 文件绝不离开你的本地网络", "getApps": "为你的其他设备获取 AllDrop", "version": "版本", "perSec": "/秒", "left": "剩余 {t}", "justNow": "刚刚", "minAgo": "{n} 分钟前", "hourAgo": "{n} 小时前", "sec": "秒", "min": "分", "browserTag": "浏览器", "thisHub": "网络中枢", "cantSelf": "无法发送给自己", "hidden": "列表已清除", "os_browser": "浏览器", "os_windows": "Windows", "os_mac": "Mac", "os_linux": "Linux", "os_android": "Android", "os_ios": "iPhone", "os_ipad": "iPad"}
  };

  function store(k, v) {
    try { if (v === undefined) return window.localStorage.getItem(k); window.localStorage.setItem(k, v); } catch (e) {}
    return null;
  }

  var lang = store('ad_lang') || 'en';  // English is the default; the language button switches it
  if (!L[lang]) lang = 'en';

  function t(k, a) {
    var s = L[lang][k] || L.en[k] || k;
    if (a) { for (var p in a) { if (a.hasOwnProperty(p)) s = s.split('{' + p + '}').join(a[p]); } }
    return s;
  }
  function $(id) { return document.getElementById(id); }
  function esc(s) {
    return String(s == null ? '' : s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }
  function show(el, on) {
    if (typeof el === 'string') el = $(el);
    if (!el) return;
    if (on) el.className = el.className.replace(/\s*\bhidden\b/g, '');
    else if (!/\bhidden\b/.test(el.className)) el.className += ' hidden';
  }
  function now() { return new Date().getTime(); }
  // singular forms where the language has one ("1 file", not "1 files")
  function cnt(k, n) { return (n === 1 && L[lang][k + '1']) ? L[lang][k + '1'] : t(k, { n: n }); }
  function num(s) { return lang === 'tr' ? s.replace('.', ',') : s; }
  function fmtSize(n) {
    n = n || 0;
    if (n < 1024) return n + ' B';
    var u = ['KB', 'MB', 'GB', 'TB'], i = -1;
    do { n /= 1024; i++; } while (n >= 1024 && i < 3);
    return num(n >= 100 ? n.toFixed(0) : n.toFixed(1)) + ' ' + u[i];
  }
  function fmtDur(s) {
    s = Math.max(1, Math.round(s));
    if (s < 60) return s + ' ' + t('sec');
    var m = Math.floor(s / 60);
    return m + ' ' + t('min') + ' ' + (s % 60) + ' ' + t('sec');
  }
  function ago(ts) {
    var d = (now() - ts) / 1000;
    if (d < 60) return t('justNow');
    if (d < 3600) return t('minAgo', { n: Math.floor(d / 60) });
    return t('hourAgo', { n: Math.floor(d / 3600) });
  }
  function osLabel(os) { return t('os_' + (os || 'browser')); }
  function mediaType(name) {
    return /\.(jpe?g|png|gif|webp|heic|heif|bmp|svg|mp4|mov|m4v|webm|mp3|m4a|aac|wav|ogg|pdf|txt)$/i.test(name || '');
  }

  var ICON = {
    folder: '<svg viewBox="0 0 24 24"><path fill="currentColor" d="M3 6.5A1.5 1.5 0 0 1 4.5 5h4.6l2 2H19.5A1.5 1.5 0 0 1 21 8.5v9a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 17.5v-11z"/></svg>',
    file: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M6 2h8l5 5v13a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2zm7 1.5V8h4.5L13 3.5z"/></svg>',
    image: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zm0 13.5h16l-5-6.5-3.5 4.5-2.5-3L4 17.5zM8 10a2 2 0 1 0 0-4 2 2 0 0 0 0 4z"/></svg>',
    lock: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M12 2a5 5 0 0 1 5 5v3h1a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-8a2 2 0 0 1 2-2h1V7a5 5 0 0 1 5-5zm0 2a3 3 0 0 0-3 3v3h6V7a3 3 0 0 0-3-3zm0 9a2 2 0 0 0-1 3.7V19h2v-2.3A2 2 0 0 0 12 13z"/></svg>',
    laptop: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M5.5 4h13A1.5 1.5 0 0 1 20 5.5V15H4V5.5A1.5 1.5 0 0 1 5.5 4zM6 6v7h12V6H6zM2 16.5h20l-.6 1.6a1.5 1.5 0 0 1-1.4.9H4a1.5 1.5 0 0 1-1.4-.9L2 16.5z"/></svg>',
    desktop: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M4.5 4h15A1.5 1.5 0 0 1 21 5.5v10a1.5 1.5 0 0 1-1.5 1.5H14v2h2.5v1.5h-9V19H10v-2H4.5A1.5 1.5 0 0 1 3 15.5v-10A1.5 1.5 0 0 1 4.5 4zM5 6v9h14V6H5z"/></svg>',
    phone: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M8 2h8a2 2 0 0 1 2 2v16a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2zm0 2v14h8V4H8zm4 15a1 1 0 1 0 0 2 1 1 0 0 0 0-2z"/></svg>',
    tablet: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M5 2h14a2 2 0 0 1 2 2v16a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2zm0 2v14.5h14V4H5zm7 15.3a.9.9 0 1 0 0 1.8.9.9 0 0 0 0-1.8z"/></svg>',
    globe: '<svg viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm6.9 9h-3a15 15 0 0 0-1.3-5.6A8 8 0 0 1 18.9 11zM12 4c.9 1.2 1.8 3.6 2 7h-4c.2-3.4 1.1-5.8 2-7zM9.4 5.4A15 15 0 0 0 8.1 11h-3a8 8 0 0 1 4.3-5.6zM5.1 13h3a15 15 0 0 0 1.3 5.6A8 8 0 0 1 5.1 13zm6.9 7c-.9-1.2-1.8-3.6-2-7h4c-.2 3.4-1.1 5.8-2 7zm2.6-1.4a15 15 0 0 0 1.3-5.6h3a8 8 0 0 1-4.3 5.6z"/></svg>'
  };
  function iconFor(os) {
    switch (os) {
      case 'mac': return ICON.laptop;
      case 'windows': case 'linux': return ICON.desktop;
      case 'android': case 'ios': return ICON.phone;
      case 'ipad': return ICON.tablet;
    }
    return ICON.globe;
  }

  function detectOS() {
    var ua = navigator.userAgent || '';
    if (/iPad/.test(ua) || (/Macintosh/.test(ua) && navigator.maxTouchPoints > 1)) return 'ipad';
    if (/iPhone|iPod/.test(ua)) return 'ios';
    if (/Android/.test(ua)) return 'android';
    if (/Windows/.test(ua)) return 'windows';
    if (/Macintosh|Mac OS X/.test(ua)) return 'mac';
    if (/Linux|CrOS/.test(ua)) return 'linux';
    return 'browser';
  }

  // ---------- server calls ----------
  function api(method, url, body, cb) {
    var x = new XMLHttpRequest();
    x.open(method, url, true);
    x.setRequestHeader('X-AD', '1');
    if (body != null) x.setRequestHeader('Content-Type', 'application/json');
    x.onreadystatechange = function () {
      if (x.readyState !== 4) return;
      var d = null;
      try { d = JSON.parse(x.responseText); } catch (e) {}
      if (cb) cb(x.status, d || {});
    };
    x.send(body != null ? JSON.stringify(body) : null);
  }

  // Browsers can pick whole folders (webkitdirectory) on computers; phones use the native app pickers.
  var os = detectOS();
  var dirOK = ('webkitdirectory' in document.createElement('input')) && os !== 'ios' && os !== 'ipad' && os !== 'android';
  var S = null, pollT = null, wasOffline = false, locked = false, helloSent = false;
  var sends = [], seq = 0, picked = [], pendingDev = null, hiddenSends = {}, seenDone = {};
  var modalKind = null, modalOffer = null, dismissed = {}, urlIdx = 0, lastQR = '', lastDevKey = '';

  function poll() {
    clearTimeout(pollT);
    api('GET', '/api/state', null, function (code, d) {
      if (code === 200) {
        if (wasOffline) { wasOffline = false; show('offline', false); }
        S = d;
        if (!d.me.owner && !helloSent) sayHello();
        render();
      } else if (code === 401) {
        lock(t('lockTitle'), t('lockText'));
      } else if (!locked) {
        wasOffline = true;
        $('offline').innerHTML = esc(t('offline'));
        show('offline', true);
      }
      if (!locked) pollT = setTimeout(poll, (document.hidden && !busy()) ? 4000 : 1000);
    });
  }

  function myGuestName() {
    var n = store('ad_name');
    if (!n) {
      var chars = 'ABCDEFGHJKLMNPRSTUVYZ23456789';
      n = osLabel(detectOS()) + ' ' + chars.charAt(Math.floor(Math.random() * chars.length)) + chars.charAt(Math.floor(Math.random() * chars.length));
      store('ad_name', n);
    }
    return n;
  }
  function sayHello() {
    helloSent = true;
    api('POST', '/api/hello', { name: myGuestName(), os: detectOS() }, function () {});
  }

  function lock(title, text) {
    locked = true;
    clearTimeout(pollT);
    show('bar', false);
    show('settingsBtn', false);
    closeModal();
    $('main').innerHTML = '<div class="card lock"><div class="avatar k-guest">' + ICON.phone + '</div><h2>' + esc(title) + '</h2><p class="hint">' + esc(text) + '</p></div>';
  }

  // ---------- rendering ----------
  function applyStatic() {
    document.documentElement.setAttribute('lang', lang);
    document.documentElement.setAttribute('dir', lang === 'ar' ? 'rtl' : 'ltr');
    if ($('langSel')) { $('langSel').value = lang; $('langSel').title = t('language'); }
    $('settingsBtn').title = t('settings');
    $('tDevices').innerHTML = esc(t('devices'));
    $('tTransfers').innerHTML = esc(t('transfers'));
    $('tInbox').innerHTML = esc(t('inbox'));
    $('zipAll').innerHTML = esc(t('zipAll'));
    $('clearInbox').innerHTML = esc(t('clear'));
    $('tJoin').innerHTML = esc(t('join'));
    $('tJoinSub').innerHTML = esc(t('joinSub'));
    $('tJoin1').innerHTML = esc(t('join1'));
    $('tJoin2').innerHTML = esc(t('join2'));
    $('tJoin3').innerHTML = esc(t('join3'));
    $('nextUrl').innerHTML = esc(t('otherNet'));
    $('tHistory').innerHTML = esc(t('history'));
    $('openFolder').innerHTML = esc(t('openFolder'));
    $('pickFiles').innerHTML = esc(t('pickFiles'));
    $('pickDir').innerHTML = esc(t('pickDir'));
    $('pickMedia').innerHTML = esc(t('pickMedia'));
    $('dropText').innerHTML = esc(t('dropHere'));
    lastDevKey = '';
  }

  function render() {
    if (!S || locked) return;
    var owner = S.me.owner;
    $('meName').innerHTML = esc(S.me.name || '');
    show('settingsBtn', owner);
    show('scanBtn', owner);
    $('scanBtn').innerHTML = esc(S.host.scanning ? t('scanning') : t('scan'));
    renderDevices();
    $('hint').innerHTML = esc(owner && !S.host.mobile ? t('hintOwner') : t('hintGuest'));
    renderTray();
    checkAutoSend();
    renderTransfers();
    renderInbox();
    renderJoin();
    renderHistory();
    renderOffers();
    $('foot').innerHTML = esc(t('secure')) + ' · AllDrop ' + esc(S.host.version || '');
    var _ga = $('getApps'); if (_ga) _ga.innerHTML = esc(t('getApps'));
  }

  function renderDevices() {
    var devs = S.devices || [];
    var key = lang + JSON.stringify(devs);
    if (key === lastDevKey) return;
    lastDevKey = key;
    var h = '';
    for (var i = 0; i < devs.length; i++) {
      var d = devs[i];
      var sub = osLabel(d.os);
      if (d.kind === 'guest') sub += ' · ' + t('browserTag');
      h += '<div class="device" data-dev="' + esc(d.id) + '" role="button" tabindex="0">' +
        '<div class="avatar k-' + esc(d.kind) + '">' + iconFor(d.os) + '</div>' +
        '<div class="dname" title="' + esc(d.name) + '">' + esc(d.name) + '</div>' +
        '<div class="dsub">' + esc(sub) + '</div></div>';
    }
    $('devices').innerHTML = h;
    show('empty', devs.length === 0);
    $('emptyText').innerHTML = esc(t('emptyOwner'));
  }

  function findDev(id) {
    var devs = (S && S.devices) || [];
    for (var i = 0; i < devs.length; i++) if (devs[i].id === id) return devs[i];
    return null;
  }

  function sumSize(items) { var n = 0; for (var i = 0; i < items.length; i++) n += items[i].f.size; return n; }

  // ---------- the "to send" tray: any number of files and folders ----------
  function keyOf(it) { return (it.rel || it.f.name) + '|' + it.f.size + '|' + (it.f.lastModified || 0); }
  function addPicked(items) {
    var seen = {};
    for (var i = 0; i < picked.length; i++) seen[keyOf(picked[i])] = 1;
    for (var j = 0; j < items.length; j++) {
      var k = keyOf(items[j]);
      if (!seen[k]) { seen[k] = 1; picked.push(items[j]); }
    }
    renderTray();
  }
  function pickedGroups() {
    var out = [], idx = {};
    for (var i = 0; i < picked.length; i++) {
      var it = picked[i], rel = it.rel || '', slash = rel.indexOf('/');
      var dir = slash > 0, name = dir ? rel.substring(0, slash) : it.f.name;
      var key = (dir ? 'd:' : 'f:') + (dir ? name : keyOf(it));
      if (idx[key] === undefined) { idx[key] = out.length; out.push({ key: key, name: name, dir: dir, count: 0, total: 0, src: 'p' }); }
      var g = out[idx[key]];
      g.count++;
      g.total += it.f.size;
    }
    return out;
  }
  function stagedCount() { return (S && S.me && S.me.owner && S.staged) ? S.staged.count : 0; }
  function trayCount() { return picked.length + stagedCount(); }

  function renderTray() {
    if (!S) return;
    show('stagedBox', false);
    var groups = pickedGroups();
    var sg = (S.me.owner && S.staged && S.staged.groups) || [];
    for (var i = 0; i < sg.length; i++) groups.push({ key: (sg[i].dir ? 'd:' : 'f:') + sg[i].name, name: sg[i].name, dir: sg[i].dir, count: sg[i].count, total: sg[i].total, src: 's' });
    var n = trayCount();
    show('pickedBox', n > 0);
    if (!n) return;
    var files = 0, folders = 0, total = 0;
    for (var k = 0; k < groups.length; k++) { files += groups[k].count; total += groups[k].total; if (groups[k].dir) folders++; }
    var summary = (folders ? cnt('nFolders', folders) + ', ' : '') + cnt('nFiles', files) + ' · ' + fmtSize(total);
    var rows = '';
    for (var r = 0; r < groups.length && r < 60; r++) {
      var g = groups[r];
      var rm = g.src === 'p' ? 'data-act="rmPicked" data-id="' + esc(g.key) + '"' : 'data-act="rmStaged" data-id="' + esc(g.name) + '" data-dir="' + (g.dir ? '1' : '0') + '"';
      rows += '<div class="trow"><span class="ticon' + (g.dir ? ' dir' : '') + '">' + (g.dir ? ICON.folder : (/\.(jpe?g|png|gif|webp|heic|heif|bmp|mp4|mov|m4v|webm)$/i.test(g.name) ? ICON.image : ICON.file)) + '</span>' +
        '<span class="tname" title="' + esc(g.name) + '">' + esc(g.name) + '</span>' +
        '<span class="tmeta">' + esc((g.dir ? cnt('nFiles', g.count) + ' · ' : '') + fmtSize(g.total)) + '</span>' +
        '<button class="x" type="button" ' + rm + ' aria-label="' + esc(t('remove')) + '">×</button></div>';
    }
    if (groups.length > 60) rows += '<div class="trow more">' + esc(t('more', { n: groups.length - 60 })) + '</div>';
    $('pickedBox').innerHTML = '<div class="tray-head"><div><b>' + esc(t('trayTitle')) + '</b><small>' + esc(summary) + '</small></div>' +
      '<button class="btn small" type="button" data-act="clearTray">' + esc(t('clear')) + '</button></div>' +
      '<div class="tray-list">' + rows + '</div><p class="hint">' + esc(t('trayHint')) + '</p>';
  }
  function renderStaged() { renderTray(); }

  // ---------- choosing what to send ----------
  var NATIVE = !!(window.AllDropNative && typeof window.AllDropNative.pick === 'function');
  var autoSend = null; // native pickers stage files on the engine; send them to this device when they arrive
  function pickKinds() {
    var kinds = ['files'];
    if (dirOK || NATIVE) kinds.push('folder');
    if (NATIVE) kinds.push('media');
    return kinds;
  }
  function pick(kind, dev) {
    pendingDev = dev || null;
    autoSend = null;
    if (NATIVE) {
      if (dev) autoSend = { id: dev.id, t: now(), base: stagedCount() };
      try { window.AllDropNative.pick(kind); } catch (e) {}
      return;
    }
    var input = kind === 'folder' ? $('dirPick') : (kind === 'media' ? $('mediaPick') : $('filePick'));
    input.value = '';
    input.click();
  }
  function openChooser(dev) {
    var kinds = pickKinds();
    if (kinds.length === 1) { pick('files', dev); return; }
    modalKind = 'choose';
    var labels = { files: ['chooseFiles', ICON.file], folder: ['chooseFolder', ICON.folder], media: ['chooseMedia', ICON.image] };
    var btns = '';
    for (var i = 0; i < kinds.length; i++) {
      btns += '<button class="choice" type="button" data-act="pickFor" data-kind="' + kinds[i] + '" data-id="' + esc(dev.id) + '">' +
        '<span class="cicon">' + labels[kinds[i]][1] + '</span><span>' + esc(t(labels[kinds[i]][0])) + '</span></button>';
    }
    openModal('<div class="center"><div class="avatar big-avatar k-' + esc(dev.kind) + '">' + iconFor(dev.os) + '</div>' +
      '<h3>' + esc(t('chooseTitle', { name: dev.name })) + '</h3><p>' + esc(t('chooseHint')) + '</p></div>' +
      '<div class="choices">' + btns + '</div>' +
      '<div class="row"><button class="btn" type="button" data-act="close">' + esc(t('cancel')) + '</button></div>');
  }
  function checkAutoSend() {
    if (!autoSend || !S) return;
    if (now() - autoSend.t > 10 * 60 * 1000) { autoSend = null; return; }
    if (stagedCount() > autoSend.base) {
      var dev = findDev(autoSend.id);
      autoSend = null;
      if (dev) sendTray(dev);
    }
  }

  function badge(status) {
    var cls = status === 'done' ? 'ok' : (status === 'pending' || status === 'starting' || status === 'needCode') ? 'wait' :
      (status === 'error' || status === 'declined' || status === 'expired' || status === 'canceled') ? 'bad' : '';
    return '<span class="badge ' + cls + '">' + esc(t(status === 'needCode' ? 'needCode' : status)) + '</span>';
  }
  function active(st) { return st === 'starting' || st === 'pending' || st === 'needCode' || st === 'sending' || st === 'accepted'; }

  function trRow(key, title, sub, status, done, total, speed, cancelAct, codeAct) {
    var pct = total > 0 ? Math.min(100, Math.floor(done * 100 / total)) : (status === 'done' ? 100 : 0);
    var cls = status === 'done' ? ' ok' : (status === 'error' || status === 'declined' || status === 'expired' || status === 'canceled') ? ' bad' : '';
    var meta = fmtSize(done) + ' / ' + fmtSize(total);
    if (speed > 0 && active(status)) {
      meta += ' · ' + fmtSize(speed) + t('perSec');
      if (total > done) meta += ' · ' + t('left', { t: fmtDur((total - done) / speed) });
    }
    var btn = active(status) ? '<button class="x" type="button" data-act="' + cancelAct + '" data-id="' + esc(key) + '">' + esc(t('cancel')) + '</button>' :
      '<button class="x" type="button" data-act="hide" data-id="' + esc(key) + '">×</button>';
    if (codeAct) btn = '<button class="btn small primary" type="button" data-act="enterCode" data-id="' + esc(key) + '">' + esc(t('codeEnter')) + '</button>' + btn;
    return '<div class="tr' + cls + '"><div class="tr-row"><div class="tr-title">' + esc(title) + badge(status) + '</div>' + btn + '</div>' +
      '<div class="bar"><i style="width:' + pct + '%"></i></div><div class="tr-meta">' + esc(sub) + ' · ' + esc(meta) + '</div></div>';
  }

  var lastTrRender = 0, trTimer = null;
  function renderTransfers(force) {
    if (!S) return;
    var nowT = now();
    if (!force && nowT - lastTrRender < 200) {
      if (!trTimer) trTimer = setTimeout(function () { trTimer = null; renderTransfers(true); }, 220);
      return;
    }
    lastTrRender = nowT;
    var h = '';
    var inc = S.incoming || [];
    for (var j = 0; j < inc.length; j++) {
      var o = inc[j];
      if (o.status === 'pending' || hiddenSends['o' + o.id]) continue;
      var first = o.files && o.files[0] ? o.files[0].name : '';
      var title = o.count > 1 ? first + ' +' + (o.count - 1) : first;
      h += trRow('o' + o.id, '← ' + title, t('from', { name: o.from.name }), o.status, o.received, o.total, speedOf('o' + o.id, o.received), 'cancelIn');
      if (o.status === 'done' && !seenDone[o.id]) { seenDone[o.id] = 1; toast(t('receivedN', { n: o.count })); }
    }
    for (var i = 0; i < sends.length; i++) {
      var s = sends[i];
      if (hiddenSends['c' + s.lid]) continue;
      var nm = s.metas.length > 1 ? s.metas[0].name + ' +' + (s.metas.length - 1) : s.metas[0].name;
      var sub = t('to', { name: s.toName }) + (s.err ? ' · ' + s.err : '');
      var needC = s.status === 'pending' && s.mode === 'code';
      h += trRow('c' + s.lid, '→ ' + nm, sub, needC ? 'needCode' : s.status, s.sent || 0, s.total, s.speed || 0, 'cancelOut', needC);
    }
    var ss = S.sends || [];
    for (var k = 0; k < ss.length; k++) {
      var v = ss[k];
      if (hiddenSends['s' + v.id]) continue;
      var st = v.status === 'accepted' ? 'sending' : v.status;
      var nm2 = v.count > 1 ? v.first + ' +' + (v.count - 1) : v.first;
      var needS = v.status === 'pending' && v.mode === 'code';
      h += trRow('s' + v.id, '→ ' + nm2, t('to', { name: v.toName }) + (v.err ? ' · ' + v.err : ''), needS ? 'needCode' : st, v.sent, v.total, speedOf('s' + v.id, v.sent), 'cancelSrv', needS);
    }
    $('transfers').innerHTML = h;
    show('transfersBox', h !== '');
    wakeLock(busy());
  }

  var speeds = {};
  function speedOf(key, bytes) {
    var s = speeds[key], tt = now();
    if (!s) { speeds[key] = { b: bytes, t: tt, v: 0 }; return 0; }
    var dt = (tt - s.t) / 1000;
    if (dt >= 0.8) {
      var inst = (bytes - s.b) / dt;
      s.v = s.v ? s.v * 0.5 + inst * 0.5 : inst;
      s.b = bytes; s.t = tt;
    }
    return s.v;
  }

  var expanded = {};
  function renderInbox() {
    var items = (!S.me.owner && S.inbox) ? S.inbox.slice(0) : [];
    show('inboxBox', items.length > 0);
    if (!items.length) return;
    var key = lang + JSON.stringify(items) + JSON.stringify(expanded);
    if (key === renderInbox.last) return;
    renderInbox.last = key;
    items.sort(function (a, b) { return b.created - a.created || (a.name < b.name ? -1 : 1); });
    var order = [], byOffer = {};
    for (var i = 0; i < items.length; i++) {
      var o = items[i].offer;
      if (!byOffer[o]) { byOffer[o] = []; order.push(o); }
      byOffer[o].push(items[i]);
    }
    var h = '';
    for (var g = 0; g < order.length; g++) {
      var list = byOffer[order[g]], size = 0;
      list.sort(function (a, b) { return a.name < b.name ? -1 : 1; });
      for (var j = 0; j < list.length; j++) size += list[j].size;
      h += '<div class="ogroup"><div class="ohead"><div class="txt"><b>' + esc(t('offerHead', { name: list[0].from, n: list.length, size: fmtSize(size) })) + '</b></div>' +
        (list.length > 1 ? '<a class="btn small primary" href="/api/inbox/zip?offer=' + encodeURIComponent(order[g]) + '">' + esc(t('zipOffer')) + '</a>' : '') + '</div>';
      var limit = expanded[order[g]] ? list.length : 5;
      for (var k = 0; k < list.length && k < limit; k++) {
        var it = list[k];
        var u = '/api/inbox/file?id=' + encodeURIComponent(it.id);
        h += '<div class="item"><div class="txt"><div class="name" title="' + esc(it.name) + '">' + esc(it.name) + '</div>' +
          '<div class="meta">' + esc(fmtSize(it.size)) + '</div></div><div class="acts">' +
          (mediaType(it.name) ? '<a class="btn small" target="_blank" href="' + u + '&inline=1">' + esc(t('open')) + '</a>' : '') +
          '<a class="btn small' + (list.length > 1 ? '' : ' primary') + '" href="' + u + '" download>' + esc(t('download')) + '</a>' +
          '<button class="btn small" type="button" data-act="rmInbox" data-id="' + esc(it.id) + '">' + esc(t('remove')) + '</button></div></div>';
      }
      if (list.length > limit) h += '<button class="link-btn more-btn" type="button" data-act="expand" data-id="' + esc(order[g]) + '">' + esc(t('showAll', { n: list.length })) + '</button>';
      h += '</div>';
    }
    $('inbox').innerHTML = h;
    show('zipAll', order.length > 1);
  }

  function renderJoin() {
    var urls = (S.me.owner && S.host.urls) || [];
    show('joinBox', urls.length > 0);
    if (!urls.length) return;
    if (urlIdx >= urls.length) urlIdx = 0;
    var u = urls[urlIdx];
    if (u !== lastQR) {
      lastQR = u;
      $('qrImg').src = '/api/qr?i=' + urlIdx + '&v=' + encodeURIComponent(u);
      $('joinUrl').innerHTML = esc(u);
    }
    show('nextUrl', urls.length > 1);
  }

  function renderHistory() {
    var hs = (S.me.owner && S.history) || [];
    show('historyBox', S.me.owner);
    if (!S.me.owner) return;
    var key = lang + JSON.stringify(hs) + Math.floor(now() / 30000);
    if (key === renderHistory.last) return;
    renderHistory.last = key;
    var h = '', rows = [], gidx = {};
    for (var i = 0; i < hs.length; i++) {
      var it = hs[i];
      if (it.group) {
        var gk = it.offer + '|' + it.group;
        if (gidx[gk] === undefined) {
          gidx[gk] = rows.length;
          rows.push({ dir: true, path: it.group, name: it.group.replace(/[\\\/]+$/, '').split(/[\\\/]/).pop(), from: it.from, ts: it.ts, size: 0, count: 0 });
        }
        var r0 = rows[gidx[gk]];
        r0.size += it.size;
        r0.count++;
      } else {
        rows.push({ dir: false, path: it.path, name: it.name, from: it.from, ts: it.ts, size: it.size });
      }
    }
    for (var j = 0; j < rows.length && j < 40; j++) {
      var row = rows[j];
      var ts = row.ts || now();
      h += '<div class="item"><div class="txt"><div class="name" title="' + esc(row.path) + '">' + (row.dir ? '<span class="ticon dir inline">' + ICON.folder + '</span>' : '') + esc(row.name) + '</div>' +
        '<div class="meta">' + esc((row.dir ? cnt('nFiles', row.count) + ' · ' : '') + fmtSize(row.size)) + ' · ' + esc(t('from', { name: row.from })) + ' · ' + esc(ago(ts)) + '</div></div>' +
        '<div class="acts"><button class="btn small" type="button" data-act="' + (row.dir && S.host.mobile ? 'openDownloads' : 'openFile') + '" data-path="' + esc(row.path) + '">' + esc(t('open')) + '</button>' +
        (S.host.mobile ? '' : '<button class="btn small" type="button" data-act="reveal" data-path="' + esc(row.path) + '">' + esc(t('show')) + '</button>') + '</div></div>';
    }
    if (!h) h = '<p class="hint">' + esc(t('none')) + ' · ' + esc(S.host.downloadDir || '') + '</p>';
    $('history').innerHTML = h;
  }

  // ---------- incoming request dialog ----------
  function renderOffers() {
    var inc = S.incoming || [];
    var pending = null;
    for (var i = 0; i < inc.length; i++) {
      if (inc[i].status === 'pending' && !dismissed[inc[i].id]) { pending = inc[i]; break; }
    }
    if (modalKind === 'offer') {
      var still = false;
      for (var j = 0; j < inc.length; j++) if (inc[j].id === modalOffer && inc[j].status === 'pending') still = true;
      if (!still) closeModal();
    }
    if (pending && modalKind !== 'offer') { openOffer(pending); return; }
    if (!modalKind) {
      var c = nextCodePrompt();
      if (c) openCode(c);
    } else if (modalKind === 'code' && !codeTargetActive(modalSend)) closeModal();
  }

  // ---------- sender: type the receiver's security code ----------
  var codeDismissed = {}, modalSend = null;
  function codeTargets() {
    var out = [];
    for (var i = 0; i < sends.length; i++) {
      var s = sends[i];
      if (s.status === 'pending' && s.mode === 'code') out.push({ key: 'c' + s.lid, id: s.id, name: s.toName, left: s.left });
    }
    var ss = (S && S.sends) || [];
    for (var k = 0; k < ss.length; k++) {
      if (ss[k].status === 'pending' && ss[k].mode === 'code') out.push({ key: 's' + ss[k].id, id: ss[k].id, name: ss[k].toName, left: ss[k].left });
    }
    return out;
  }
  function codeTargetActive(key) {
    var ts = codeTargets();
    for (var i = 0; i < ts.length; i++) if (ts[i].key === key) return ts[i];
    return null;
  }
  function nextCodePrompt() {
    var ts = codeTargets();
    for (var i = 0; i < ts.length; i++) if (!codeDismissed[ts[i].key]) return ts[i];
    return null;
  }
  function openCode(c) {
    modalKind = 'code';
    modalSend = c.key;
    var h = '<div class="center"><div class="avatar big-avatar k-host">' + ICON.lock + '</div>' +
      '<h3>' + esc(t('codeTitle', { name: c.name })) + '</h3><p>' + esc(t('codeHint')) + '</p>' +
      '<input type="text" id="codeInput" class="otp-input" inputmode="numeric" pattern="[0-9 ]*" autocomplete="one-time-code" maxlength="7" placeholder="••• •••">' +
      '<p id="codeErr" class="err"></p></div>' +
      '<div class="row"><button class="btn" type="button" data-act="codeLater" data-id="' + esc(c.key) + '">' + esc(t('close')) + '</button>' +
      '<button class="btn primary" type="button" id="codeBtn" data-act="codeSubmit" data-id="' + esc(c.key) + '">' + esc(t('codeSend')) + '</button></div>';
    openModal(h);
    try { $('codeInput').focus(); } catch (e) {}
  }
  function submitCode(key) {
    var c = codeTargetActive(key);
    var inp = $('codeInput');
    if (!c || !inp) return;
    var code = inp.value.replace(/[^0-9]/g, '');
    if (code.length !== 6) { $('codeErr').innerHTML = esc(t('codeTitle', { name: c.name })); return; }
    $('codeBtn').disabled = true;
    api('POST', '/api/send/code?id=' + encodeURIComponent(c.id), { code: code }, function (st, d) {
      var btn = $('codeBtn');
      if (btn) btn.disabled = false;
      if (st === 200 && d.ok) {
        closeModal();
        for (var i = 0; i < sends.length; i++) if ('c' + sends[i].lid === key) { sends[i].mode = ''; checkNow(sends[i]); }
        poll();
        return;
      }
      if (d.status && d.status !== 'pending') {
        closeModal();
        for (var j = 0; j < sends.length; j++) if ('c' + sends[j].lid === key) { sends[j].status = d.status; }
        renderTransfers(true);
        poll();
        return;
      }
      if ($('codeErr')) $('codeErr').innerHTML = esc(st === 200 ? t('codeWrong', { n: d.left }) : (d.error || t('error')));
      if ($('codeInput')) { $('codeInput').value = ''; $('codeInput').focus(); }
    });
  }

  function openOffer(o) {
    modalKind = 'offer';
    modalOffer = o.id;
    var files = '';
    for (var i = 0; i < o.files.length; i++) files += '<div>' + esc(o.files[i].path || o.files[i].name) + ' · ' + esc(fmtSize(o.files[i].size)) + '</div>';
    if (o.count > o.files.length) files += '<div>… +' + (o.count - o.files.length) + '</div>';
    var head = '<div class="center"><div class="avatar big-avatar k-peer">' + iconFor(o.from.os) + '</div>' +
      '<h3>' + esc(t('offerTitle', { name: o.from.name })) + '</h3><p>' + esc(t('offerSub', { n: o.count, size: fmtSize(o.total) })) + '</p></div>';
    var h;
    if (o.mode === 'code') {
      h = head + '<div class="otp" aria-live="polite">' + esc(o.code.substring(0, 3) + ' ' + o.code.substring(3)) + '</div>' +
        '<p class="center">' + esc(t('codeShow', { name: o.from.name })) + '</p>' +
        '<div class="files">' + files + '</div>' +
        (S.me.owner ? '<label class="chk"><input type="checkbox" id="trustCode" data-id="' + esc(o.id) + '"' + (o.trust ? ' checked' : '') + '> ' + esc(t('codeTrust')) + '</label>' : '') +
        '<div class="row"><button class="btn" type="button" data-act="decline" data-id="' + esc(o.id) + '">' + esc(t('decline')) + '</button></div>';
    } else {
      h = head + '<div class="files">' + files + '</div>' +
        (S.me.owner ? '<label class="chk"><input type="checkbox" id="trustChk"> ' + esc(t('trust')) + '</label>' : '') +
        '<div class="row"><button class="btn" type="button" data-act="decline" data-id="' + esc(o.id) + '">' + esc(t('decline')) + '</button>' +
        '<button class="btn primary" type="button" data-act="accept" data-id="' + esc(o.id) + '">' + esc(t('accept')) + '</button></div>';
    }
    openModal(h);
    try { if (window.navigator.vibrate) navigator.vibrate(120); } catch (e) {}
  }

  function respond(id, accept) {
    var trust = $('trustChk') && $('trustChk').checked;
    dismissed[id] = 1;
    closeModal();
    api('POST', '/api/offer/respond', { id: id, accept: accept, trust: !!trust }, function () { poll(); });
  }

  // ---------- modal ----------
  // ---------- get the app on other devices ----------
  var DL_BASE = 'https://github.com/pilotmemorable/alldrop/releases/latest/download/';
  var DL_PAGE = 'https://pilotmemorable.github.io/alldrop/';
  var DL_APPS = [
    { icon: 'windows', name: 'Windows', sub: '64-bit · .exe', url: DL_BASE + 'AllDrop-Windows-64bit.exe' },
    { icon: 'mac', name: 'macOS', sub: '.dmg · Intel & Apple Silicon', url: DL_BASE + 'AllDrop-macOS.dmg' },
    { icon: 'linux', name: 'Linux', sub: 'x64', url: DL_BASE + 'AllDrop-Linux-x64' },
    { icon: 'android', name: 'Android', sub: '.apk', url: DL_BASE + 'AllDrop-Android.apk' },
    { icon: 'ios', name: 'iPhone · iPad', sub: 'App Store', url: 'https://apps.apple.com/app/id6830793977' }
  ];
  function allLinksText() {
    var s = 'AllDrop: All Device\n' + t('dlPage') + ': ' + DL_PAGE + '\n\n';
    for (var i = 0; i < DL_APPS.length; i++) s += DL_APPS[i].name + ': ' + DL_APPS[i].url + '\n';
    return s;
  }
  function fallbackCopy(txt) {
    try {
      var ta = document.createElement('textarea');
      ta.value = txt; ta.setAttribute('readonly', ''); ta.style.position = 'fixed'; ta.style.opacity = '0';
      document.body.appendChild(ta); ta.focus(); ta.select();
      document.execCommand('copy'); document.body.removeChild(ta);
    } catch (e) {}
  }
  function copyText(txt) {
    try {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(txt).then(function () { toast(t('copied')); }, function () { fallbackCopy(txt); toast(t('copied')); });
        return;
      }
    } catch (e) {}
    fallbackCopy(txt); toast(t('copied'));
  }
  function doShareApp() {
    try {
      if (navigator.share) {
        navigator.share({ title: 'AllDrop: All Device', text: t('getAppsShareText'), url: DL_PAGE }).catch(function () {});
        return;
      }
    } catch (e) {}
    copyText(DL_PAGE + '\n\n' + allLinksText());
  }
  function openShareApp() {
    modalKind = 'share';
    var list = '';
    for (var i = 0; i < DL_APPS.length; i++) {
      var app = DL_APPS[i];
      list += '<a class="dl-row" href="' + esc(app.url) + '" target="_blank" rel="noopener">' +
        '<span class="dl-ic">' + iconFor(app.icon) + '</span>' +
        '<span class="dl-tx"><b>' + esc(app.name) + '</b><small>' + esc(app.sub) + '</small></span>' +
        '<span class="dl-go">↓</span></a>';
    }
    var h = '<h3>' + esc(t('getAppsTitle')) + '</h3>' +
      '<p class="hint">' + esc(t('getAppsSub')) + '</p>' +
      '<div class="dl-list">' + list + '</div>' +
      '<div class="dl-page"><code>' + esc(DL_PAGE) + '</code></div>' +
      '<div class="row"><button class="btn" type="button" data-act="copyLinks">' + esc(t('copyLinks')) + '</button>' +
      '<button class="btn primary" type="button" data-act="shareLink">' + esc(t('shareLink')) + '</button></div>' +
      '<div class="row"><button class="btn" type="button" data-act="close">' + esc(t('close')) + '</button></div>';
    openModal(h);
  }

  function openModal(html) {
    $('sheet').innerHTML = html;
    show('modal', true);
  }
  function closeModal() {
    modalKind = null;
    modalOffer = null;
    modalSend = null;
    show('modal', false);
  }

  function openSettings() {
    if (!S || !S.me.owner) return;
    modalKind = 'settings';
    var hst = S.host;
    var tr = '';
    var trusted = hst.trusted || [];
    for (var i = 0; i < trusted.length; i++) {
      tr += '<div><span>' + esc(trusted[i].name) + '</span><button class="link-btn" type="button" data-act="untrust" data-id="' + esc(trusted[i].id) + '">' + esc(t('remove')) + '</button></div>';
    }
    var man = hst.manual || [];
    var manual = '';
    for (var k = 0; k < man.length; k++) {
      manual += '<div><code>' + esc(man[k]) + '</code><button class="link-btn" type="button" data-act="forget" data-id="' + esc(man[k]) + '">' + esc(t('remove')) + '</button></div>';
    }
    var addr = '';
    var urls = hst.urls || [];
    for (var u = 0; u < urls.length; u++) addr += '<div><code>' + esc(urls[u].replace(/\/\?k=.*$/, '').replace('http://', '')) + '</code></div>';
    var win = hst.goos === 'windows';
    var h = '<h3>' + esc(t('settings')) + '</h3>' +
      '<div class="field"><label for="setName">' + esc(t('name')) + '</label><div class="inline"><input type="text" id="setName" maxlength="48" value="' + esc(hst.name) + '"><button class="btn" type="button" data-act="saveName">' + esc(t('save')) + '</button></div></div>' +
      (hst.mobile ? '' : '<div class="field"><label for="setDir">' + esc(t('dir')) + '</label><div class="inline"><input type="text" id="setDir" value="' + esc(hst.downloadDir) + '"><button class="btn" type="button" data-act="saveDir">' + esc(t('save')) + '</button></div>' +
        (hst.dirOK ? '' : '<small>' + esc(t('dirBad')) + '</small>') + '</div>') +
      '<div class="field"><label class="chk"><input type="checkbox" id="setCode"' + (hst.requireCode ? ' checked' : '') + '> ' + esc(t('requireCode')) + '</label><small>' + esc(t('requireCodeHint')) + '</small></div>' +
      '<div class="field"><label class="chk"><input type="checkbox" id="setAuto"' + (hst.autoAccept ? ' checked' : '') + '> ' + esc(t('autoAccept')) + '</label><small>' + esc(t('autoAcceptHint')) + '</small></div>' +
      '<div class="field"><label>' + esc(t('trusted')) + '</label><div class="list-mini">' + (tr || '<div>' + esc(t('none')) + '</div>') + '</div></div>' +
      '<hr class="sep">' +
      '<div class="field"><label for="peerAddr">' + esc(t('addPeer')) + '</label><div class="inline"><input type="text" id="peerAddr" placeholder="192.168.1.20" inputmode="decimal"><button class="btn" type="button" id="addPeerBtn" data-act="addPeer">' + esc(t('add')) + '</button></div><small>' + esc(t('addPeerHint')) + '</small>' +
      (manual ? '<div class="list-mini">' + manual + '</div>' : '') + '</div>' +
      '<div class="field"><label>' + esc(t('myAddr')) + '</label><div class="list-mini">' + addr + '</div></div>' +
      (win ? '<hr class="sep"><div class="field"><button class="btn" type="button" data-act="firewall">' + esc(t('firewall')) + '</button><small>' + esc(t('firewallHint')) + '</small></div>' +
        '<div class="field"><button class="btn" type="button" data-act="sendto">' + esc(t('sendto')) + '</button><small>' + esc(t('sendtoHint')) + '</small></div>' : '') +
      '<hr class="sep"><div class="field"><button class="btn" type="button" id="resetBtn" data-act="resetKey">' + esc(t('resetKey')) + '</button><small>' + esc(t('resetKeyHint')) + '</small></div>' +
      '<div class="row">' + (hst.mobile ? '' : '<button class="btn danger" type="button" data-act="quit">' + esc(t('quit')) + '</button>') +
      '<button class="btn primary" type="button" data-act="close">' + esc(t('close')) + '</button></div>' +
      '<p class="hint">' + esc(t('version')) + ' ' + esc(hst.version) + '</p>';
    openModal(h);
  }

  function openRename() {
    modalKind = 'rename';
    var h = '<h3>' + esc(t('rename')) + '</h3><p>' + esc(t('renameHint')) + '</p>' +
      '<input type="text" id="renameInput" maxlength="48" value="' + esc(S.me.name || '') + '">' +
      '<div class="row"><button class="btn" type="button" data-act="close">' + esc(t('cancel')) + '</button><button class="btn primary" type="button" data-act="doRename">' + esc(t('save')) + '</button></div>';
    openModal(h);
  }

  var toastT = null;
  function toast(msg) {
    var el = $('toast');
    el.innerHTML = esc(msg);
    show(el, true);
    clearTimeout(toastT);
    toastT = setTimeout(function () { show(el, false); }, 3200);
  }

  // ---------- sending ----------
  function busy() {
    for (var i = 0; i < sends.length; i++) if (active(sends[i].status)) return true;
    var inc = (S && S.incoming) || [];
    for (var j = 0; j < inc.length; j++) if (inc[j].status === 'accepted') return true;
    var ss = (S && S.sends) || [];
    for (var k = 0; k < ss.length; k++) if (active(ss[k].status)) return true;
    return false;
  }

  var wl = null, wlBusy = false;
  function wakeLock(on) {
    try {
      if (!navigator.wakeLock) return;
      if (on && !wl && !wlBusy) {
        wlBusy = true;
        navigator.wakeLock.request('screen').then(function (l) { wl = l; wlBusy = false; })['catch'](function () { wlBusy = false; });
      } else if (!on && wl) { wl.release(); wl = null; }
    } catch (e) {}
  }

  function toItems(fileList) {
    var out = [];
    for (var i = 0; i < fileList.length; i++) out.push({ f: fileList[i], rel: fileList[i].webkitRelativePath || '' });
    return out;
  }

  function tapDevice(id) {
    var dev = findDev(id);
    if (!dev) return;
    if (trayCount() > 0) { sendTray(dev); return; }
    openChooser(dev);
  }

  // Sends everything in the tray (files and folders, any number) to one device.
  function sendTray(dev) {
    if (stagedCount() > 0) {
      api('POST', '/api/send/staged', { to: dev.id }, function (code, d) {
        if (code !== 200) toast(d.error || ('HTTP ' + code));
        poll();
      });
    }
    if (picked.length) {
      var items = picked;
      picked = [];
      startSend(dev, items);
    }
    renderTray();
  }

  function startSend(dev, items) {
    if (!items.length) return;
    var metas = [], total = 0;
    for (var i = 0; i < items.length; i++) {
      metas.push({ name: items[i].f.name, path: items[i].rel || '', size: items[i].f.size });
      total += items[i].f.size;
    }
    var tr = { lid: ++seq, toName: dev.name, items: items, metas: metas, total: total, loaded: [], sent: 0, status: 'starting', xhrs: {}, speed: 0 };
    sends.unshift(tr);
    if (sends.length > 12) sends.length = 12;
    renderTransfers(true);
    api('POST', '/api/send', { to: dev.id, files: metas }, function (code, d) {
      if (code !== 200) { tr.status = 'error'; tr.err = d.error || ('HTTP ' + code); renderTransfers(true); return; }
      tr.id = d.id;
      tr.status = 'pending';
      renderTransfers(true);
      waitAccept(tr);
    });
  }

  function waitAccept(tr) {
    if (tr.status !== 'pending') return;
    api('GET', '/api/send/status?id=' + encodeURIComponent(tr.id), null, function (code, d) {
      if (tr.status !== 'pending') return;
      if (code !== 200) { tr.status = 'error'; tr.err = d.error || ('HTTP ' + code); renderTransfers(true); return; }
      tr.mode = d.mode; tr.left = d.left;
      if (d.status === 'accepted') { tr.status = 'sending'; if (modalKind === 'code' && modalSend === 'c' + tr.lid) closeModal(); renderTransfers(true); upload(tr); }
      else if (d.status === 'pending') { if (tr.mode === 'code') renderOffers(); renderTransfers(); tr.timer = setTimeout(function () { waitAccept(tr); }, 700); }
      else { tr.status = d.status; renderTransfers(true); renderOffers(); }
    });
  }
  function checkNow(tr) { clearTimeout(tr.timer); waitAccept(tr); }

  function progress(tr) {
    var b = 0;
    for (var i = 0; i < tr.metas.length; i++) b += tr.loaded[i] || 0;
    tr.sent = b;
    var tt = now();
    if (!tr.st) tr.st = { b: b, t: tt };
    var dt = (tt - tr.st.t) / 1000;
    if (dt >= 0.5) {
      var inst = (b - tr.st.b) / dt;
      tr.speed = tr.speed ? tr.speed * 0.6 + inst * 0.4 : inst;
      tr.st = { b: b, t: tt };
    }
    renderTransfers();
  }

  function upload(tr) {
    var next = 0, running = 0, finished = 0, n = tr.items.length;
    function failAll(msg) {
      if (tr.status !== 'sending') return;
      tr.status = 'error';
      tr.err = msg;
      abortAll(tr);
      api('POST', '/api/send/cancel?id=' + encodeURIComponent(tr.id), {}, function () {});
      renderTransfers(true);
    }
    function pump() {
      while (running < 3 && next < n && tr.status === 'sending') {
        (function (i) {
          running++;
          var x = new XMLHttpRequest();
          tr.xhrs[i] = x;
          x.open('POST', '/api/send/file?id=' + encodeURIComponent(tr.id) + '&i=' + i, true);
          x.setRequestHeader('X-AD', '1');
          x.setRequestHeader('Content-Type', 'application/octet-stream');
          if (x.upload) x.upload.onprogress = function (e) { tr.loaded[i] = e.loaded; progress(tr); };
          x.onreadystatechange = function () {
            if (x.readyState !== 4) return;
            delete tr.xhrs[i];
            running--;
            if (x.status === 200) {
              tr.loaded[i] = tr.metas[i].size;
              finished++;
              progress(tr);
              if (finished === n) {
                tr.status = 'done';
                tr.speed = 0;
                renderTransfers(true);
                toast(t('sentOk', { name: tr.toName }));
              } else pump();
            } else if (tr.status === 'sending') {
              var msg = 'HTTP ' + x.status;
              try { msg = JSON.parse(x.responseText).error || msg; } catch (e) {}
              if (x.status === 0) msg = t('offline');
              failAll(msg);
            }
          };
          x.send(tr.items[i].f);
        })(next++);
      }
    }
    pump();
  }

  function abortAll(tr) {
    for (var k in tr.xhrs) { if (tr.xhrs.hasOwnProperty(k)) { try { tr.xhrs[k].abort(); } catch (e) {} } }
    tr.xhrs = {};
  }

  function cancelOut(lid) {
    for (var i = 0; i < sends.length; i++) {
      var tr = sends[i];
      if ('c' + tr.lid !== lid) continue;
      var prev = tr.status;
      tr.status = 'canceled';
      abortAll(tr);
      if (tr.id && prev !== 'starting') api('POST', '/api/send/cancel?id=' + encodeURIComponent(tr.id), {}, function () {});
    }
    renderTransfers(true);
  }

  // ---------- drag & drop (files and whole folders) ----------
  function walkEntries(entries, cb) {
    var out = [], pending = 0, fired = false;
    function check() { if (pending === 0 && !fired) { fired = true; cb(out); } }
    function visit(entry, prefix) {
      if (!entry) return;
      pending++;
      if (entry.isFile) {
        entry.file(function (f) { out.push({ f: f, rel: prefix ? prefix + f.name : '' }); pending--; check(); }, function () { pending--; check(); });
      } else if (entry.isDirectory) {
        var reader = entry.createReader();
        var readMore = function () {
          reader.readEntries(function (list) {
            if (!list.length) { pending--; check(); return; }
            for (var i = 0; i < list.length; i++) visit(list[i], prefix + entry.name + '/');
            readMore();
          }, function () { pending--; check(); });
        };
        readMore();
      } else { pending--; check(); }
    }
    for (var i = 0; i < entries.length; i++) visit(entries[i], '');
    check();
  }

  function dropped(dt, cb) {
    var items = dt.items;
    if (items && items.length && items[0].webkitGetAsEntry) {
      var entries = [];
      for (var i = 0; i < items.length; i++) {
        if (items[i].kind === 'file') { var e = items[i].webkitGetAsEntry(); if (e) entries.push(e); }
      }
      walkEntries(entries, cb);
      return;
    }
    cb(toItems(dt.files || []));
  }

  function hasFiles(e) {
    var ty = e.dataTransfer && e.dataTransfer.types;
    if (!ty) return false;
    for (var i = 0; i < ty.length; i++) if (ty[i] === 'Files') return true;
    return !!(ty.contains && ty.contains('Files'));
  }

  function devEl(el) {
    while (el && el !== document) {
      if (el.getAttribute && el.getAttribute('data-dev')) return el;
      el = el.parentNode;
    }
    return null;
  }

  var dragDepth = 0, dropTarget = null;
  function setDropTarget(el) {
    if (dropTarget === el) return;
    if (dropTarget) dropTarget.className = dropTarget.className.replace(/\s*\bdrop\b/g, '');
    dropTarget = el;
    if (el) el.className += ' drop';
    var d = el ? findDev(el.getAttribute('data-dev')) : null;
    $('dropText').innerHTML = esc(d ? t('dropOn', { name: d.name }) : t('dropHere'));
  }

  document.addEventListener('dragenter', function (e) {
    if (!hasFiles(e) || locked) return;
    dragDepth++;
    show('dropOverlay', true);
  });
  document.addEventListener('dragleave', function (e) {
    if (!hasFiles(e)) return;
    dragDepth--;
    if (dragDepth <= 0) { dragDepth = 0; show('dropOverlay', false); setDropTarget(null); }
  });
  document.addEventListener('dragover', function (e) {
    if (!hasFiles(e) || locked) return;
    e.preventDefault();
    try { e.dataTransfer.dropEffect = 'copy'; } catch (x) {}
    setDropTarget(devEl(e.target));
  });
  document.addEventListener('drop', function (e) {
    if (!hasFiles(e) || locked) return;
    e.preventDefault();
    dragDepth = 0;
    show('dropOverlay', false);
    var el = devEl(e.target);
    setDropTarget(null);
    var dev = el ? findDev(el.getAttribute('data-dev')) : null;
    dropped(e.dataTransfer, function (items) {
      if (!items.length) return;
      if (dev) startSend(dev, items);
      else addPicked(items);
    });
  });

  // ---------- clicks ----------
  function attr(el, name) {
    while (el && el !== document) {
      if (el.getAttribute && el.getAttribute(name)) return el;
      el = el.parentNode;
    }
    return null;
  }

  document.addEventListener('click', function (e) {
    var dEl = devEl(e.target);
    if (dEl) { tapDevice(dEl.getAttribute('data-dev')); return; }
    var a = attr(e.target, 'data-act');
    if (!a) {
      if (e.target === $('modal') && modalKind !== 'offer' && modalKind !== 'code') closeModal();
      return;
    }
    var act = a.getAttribute('data-act'), id = a.getAttribute('data-id'), p = a.getAttribute('data-path');
    switch (act) {
      case 'accept': respond(id, true); break;
      case 'codeSubmit': submitCode(id); break;
      case 'codeLater': codeDismissed[id] = 1; closeModal(); break;
      case 'enterCode': codeDismissed[id] = 0; var ct = codeTargetActive(id); if (ct) openCode(ct); break;
      case 'decline': respond(id, false); break;
      case 'close': closeModal(); break;
      case 'shareApp': e.preventDefault(); openShareApp(); break;
      case 'shareLink': doShareApp(); break;
      case 'copyLinks': copyText(allLinksText()); break;
      case 'hide': hiddenSends[id] = 1; renderTransfers(true); break;
      case 'cancelOut': cancelOut(id); break;
      case 'cancelIn': api('POST', '/api/offer/respond', { id: id.substring(1), accept: false }, function () { poll(); }); break;
      case 'cancelSrv': api('POST', '/api/send/cancel?id=' + encodeURIComponent(id.substring(1)), {}, function () { poll(); }); break;
      case 'clearPicked': picked = []; renderTray(); break;
      case 'clearStaged': api('POST', '/api/stage/clear', {}, function () { poll(); }); break;
      case 'clearTray':
        picked = [];
        if (stagedCount() > 0) api('POST', '/api/stage/clear', {}, function () { poll(); });
        renderTray();
        break;
      case 'rmPicked':
        var keep = [];
        for (var q = 0; q < picked.length; q++) {
          var it = picked[q], rel = it.rel || '', sl = rel.indexOf('/');
          var k2 = sl > 0 ? 'd:' + rel.substring(0, sl) : 'f:' + keyOf(it);
          if (k2 !== id) keep.push(it);
        }
        picked = keep;
        renderTray();
        break;
      case 'rmStaged': api('POST', '/api/stage/remove', { name: id, dir: a.getAttribute('data-dir') === '1' }, function () { poll(); }); break;
      case 'pickFor': var pd = findDev(id); closeModal(); pick(a.getAttribute('data-kind'), pd); break;
      case 'expand': expanded[id] = 1; renderInbox(); break;
      case 'openDownloads':
        if (window.AllDropNative && window.AllDropNative.openDownloads) window.AllDropNative.openDownloads();
        else api('POST', '/api/open', { kind: 'folder' }, function () {});
        break;
      case 'rmInbox': api('POST', '/api/inbox/remove', { id: id }, function () { poll(); }); break;
      case 'openFile':
        if (window.AllDropNative && window.AllDropNative.openFile) window.AllDropNative.openFile(p);
        else api('POST', '/api/open', { kind: 'file', path: p }, function () {});
        break;
      case 'reveal': api('POST', '/api/open', { kind: 'reveal', path: p }, function () {}); break;
      case 'saveName':
        api('POST', '/api/settings', { name: $('setName').value }, function (c) { toast(c === 200 ? t('saved') : t('error')); poll(); });
        break;
      case 'saveDir':
        api('POST', '/api/settings', { downloadDir: $('setDir').value }, function (c, d) { toast(c === 200 ? t('saved') : (d.error || t('error'))); poll(); });
        break;
      case 'untrust': api('POST', '/api/trust/remove', { id: id }, function () { setTimeout(function () { poll(); setTimeout(openSettings, 400); }, 50); }); break;
      case 'forget': api('POST', '/api/peer/forget', { addr: id }, function () { poll(); setTimeout(openSettings, 400); }); break;
      case 'addPeer':
        var btn = $('addPeerBtn');
        btn.disabled = true;
        btn.innerHTML = esc(t('adding'));
        api('POST', '/api/peer/add', { addr: $('peerAddr').value }, function (c) {
          toast(c === 200 ? t('added') : t('notFound'));
          btn.disabled = false;
          btn.innerHTML = esc(t('add'));
          poll();
        });
        break;
      case 'firewall': api('POST', '/api/win/firewall', {}, function (c, d) { toast(c === 200 ? t('fwOk') : (d.error || t('error'))); }); break;
      case 'sendto': api('POST', '/api/win/sendto', {}, function (c, d) { toast(c === 200 ? t('sendtoOk') : (d.error || t('error'))); }); break;
      case 'resetKey':
        if (a.getAttribute('data-sure') !== '1') { a.setAttribute('data-sure', '1'); a.innerHTML = esc(t('sure')); return; }
        api('POST', '/api/key/regen', {}, function () { lastQR = ''; closeModal(); poll(); });
        break;
      case 'quit':
        api('POST', '/api/quit', {}, function () { lock(t('quitTitle'), t('quitText')); });
        break;
      case 'doRename':
        var v = $('renameInput').value.replace(/^\s+|\s+$/g, '');
        if (!v) return;
        closeModal();
        if (S.me.owner) api('POST', '/api/settings', { name: v }, function () { poll(); });
        else { store('ad_name', v); api('POST', '/api/hello', { name: v, os: detectOS() }, function () { poll(); }); }
        break;
    }
  });

  document.addEventListener('change', function (e) {
    if (e.target && e.target.id === 'setAuto') {
      api('POST', '/api/settings', { autoAccept: e.target.checked }, function () { poll(); });
    }
    if (e.target && e.target.id === 'setCode') {
      api('POST', '/api/settings', { requireCode: e.target.checked }, function () { poll(); });
    }
    if (e.target && e.target.id === 'trustCode') {
      api('POST', '/api/offer/trust', { id: e.target.getAttribute('data-id'), trust: e.target.checked }, function () {});
    }
  });
  document.addEventListener('input', function (e) {
    if (e.target && e.target.id === 'codeInput') {
      var v = e.target.value.replace(/[^0-9]/g, '');
      if (v.length > 6) v = v.substring(0, 6);
      e.target.value = v.length > 3 ? v.substring(0, 3) + ' ' + v.substring(3) : v;
      if (v.length === 6) submitCode(modalSend);
    }
  });

  document.addEventListener('keydown', function (e) {
    if (e.keyCode === 27 && modalKind && modalKind !== 'offer') { if (modalKind === 'code') codeDismissed[modalSend] = 1; closeModal(); }
    if (e.keyCode === 13 && modalKind === 'code' && e.target && e.target.id === 'codeInput') { e.preventDefault(); submitCode(modalSend); }
    if ((e.keyCode === 13 || e.keyCode === 32) && e.target && e.target.getAttribute && e.target.getAttribute('data-dev')) {
      e.preventDefault();
      tapDevice(e.target.getAttribute('data-dev'));
    }
  });

  function onPicked() {
    var items = toItems(this.files || []);
    if (!items.length) return;
    if (pendingDev) { var d = pendingDev; pendingDev = null; startSend(d, items); }
    else addPicked(items);
  }
  $('filePick').onchange = onPicked;
  $('dirPick').onchange = onPicked;
  $('mediaPick').onchange = onPicked;
  $('pickFiles').onclick = function () { pick('files', null); };
  $('pickDir').onclick = function () { pick('folder', null); };
  $('pickMedia').onclick = function () { pick('media', null); };
  $('settingsBtn').onclick = openSettings;
  $('meChip').onclick = function () { if (!S) return; if (S.me.owner) openSettings(); else openRename(); };
  $('langSel').onchange = function () {
    lang = this.value;
    if (!L[lang]) lang = 'en';
    store('ad_lang', lang);
    applyStatic();
    renderInbox.last = renderHistory.last = '';
    if (modalKind === 'settings') openSettings();
    render();
  };
  $('scanBtn').onclick = function () { api('POST', '/api/scan', {}, function () { poll(); }); };
  $('nextUrl').onclick = function () { urlIdx++; lastQR = ''; render(); };
  $('openFolder').onclick = function () {
    if (window.AllDropNative && window.AllDropNative.openDownloads) window.AllDropNative.openDownloads();
    else api('POST', '/api/open', { kind: 'folder' }, function () {});
  };
  $('clearInbox').onclick = function () { api('POST', '/api/inbox/remove', { all: true }, function () { poll(); }); };

  show('pickDir', dirOK || NATIVE);
  show('pickMedia', NATIVE);

  document.addEventListener('visibilitychange', function () { if (!document.hidden && !locked) poll(); });
  window.addEventListener('beforeunload', function (e) {
    for (var i = 0; i < sends.length; i++) {
      if (sends[i].status === 'sending') { e.returnValue = 'AllDrop'; return 'AllDrop'; }
    }
  });

  applyStatic();
  poll();
})();
