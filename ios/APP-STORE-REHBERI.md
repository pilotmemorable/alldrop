# AllDrop'u App Store'a yükleme rehberi

Bu klasör iPhone ve iPad uygulamasının tamamını içerir: Swift kabuğu, aynı Go aktarım motoru (gomobile ile iOS çerçevesine derlenir), simge, gizlilik manifesti, mağaza metinleri (Türkçe ve İngilizce) ve ekran görüntüleri.

Uygulama adı: **AllDrop: All Device**
Alt başlık: **Mac, Windows, Android, iPhone** (TR) / **All Devices All Transfer** (EN)
Açıklamada: *Mac to Windows · Win to Mac · iOS to Android · Android to iOS · All Devices All Transfer*

> "AirDrop" kelimesi Apple'ın tescilli markasıdır. Ad, alt başlık, anahtar kelimeler ve açıklamada kullanılmadı; kullanılırsa inceleme reddedilir (Kural 5.2.5).

---

## 1. Apple tarafında bir kerelik hazırlık

1. **Bundle ID oluşturun:** developer.apple.com → Certificates, Identifiers & Profiles → Identifiers → **+** → App IDs → App.
   Örnek: `com.adiniz.alldrop` (benzersiz olmalı). Ek yetenek seçmenize gerek yok.
2. **App Store Connect'te uygulamayı açın:** appstoreconnect.apple.com → Uygulamalar → **+** → Yeni Uygulama
   - Platform: iOS · Ad: `AllDrop: All Device` · Birincil dil: Türkçe
   - Paket kimliği: 1. adımdaki Bundle ID · SKU: `alldrop-001`
   - Ad başkası tarafından alınmışsa küçük bir değişiklik yapın (örn. `AllDrop: All Device Transfer`).

## 2. Uygulamayı derleyip yükleme: iki yol

### Yol A: Kendi Mac'inizde (en kolay)

```bash
# Bir kere: Xcode (App Store'dan), Homebrew, Go ve XcodeGen
brew install go xcodegen

cd alldrop
./ios/build-engine.sh          # Go motorunu iOS çerçevesine derler (birkaç dakika)
cd ios && xcodegen generate    # AllDrop.xcodeproj oluşturur
open AllDrop.xcodeproj
```

Xcode'da:
1. Sol üstte **AllDrop** hedefi → *Signing & Capabilities* → **Team**: geliştirici hesabınız, **Bundle Identifier**: 1. adımdaki kimlik.
2. Önce iPhone'unuzu kabloyla bağlayıp ▶︎ ile deneyin.
3. Üst menüden **Product → Archive** → açılan pencerede **Distribute App → App Store Connect → Upload**.

### Yol B: GitHub Actions ile otomatik (Mac gerekmez)

Depo GitHub'a yüklendikten sonra **Settings → Secrets and variables → Actions** altında:

| Ad | Nereden |
|---|---|
| `APPLE_TEAM_ID` (secret) | developer.apple.com → Membership → Team ID |
| `ASC_KEY_ID` (secret) | App Store Connect → Kullanıcılar ve Erişim → Entegrasyonlar → Anahtarlar → **+** (rol: App Manager) |
| `ASC_ISSUER_ID` (secret) | Aynı sayfadaki Issuer ID |
| `ASC_KEY_P8` (secret) | İndirilen `AuthKey_XXXX.p8` dosyasının içeriği |
| `IOS_BUNDLE_ID` (variable) | 1. adımdaki Bundle ID |

Sonra **Actions → iOS app → Run workflow** (ya da `v1.0.1` gibi bir etiket gönderin). Derlenen uygulama App Store Connect'e (TestFlight) yüklenir.

## 3. Mağaza sayfası

AIRDROP klasöründeki `App Store metinleri` (fastlane biçiminde) içindeki metinleri App Store Connect'e kopyalayın:

| Alan | Dosya |
|---|---|
| Ad | `tr/name.txt` |
| Alt başlık | `tr/subtitle.txt`, `en-US/subtitle.txt` |
| Tanıtım metni | `*/promotional_text.txt` |
| Açıklama | `*/description.txt` |
| Anahtar kelimeler | `*/keywords.txt` (100 karakter sınırına uygun) |
| İnceleme notu | `review_information/notes.txt` (İngilizce, inceleme ekibi için test adımları) |

- **Ekran görüntüleri:** AIRDROP klasöründeki `App Store ekran goruntuleri` (her cihaz için 5 görüntü; ilki çoklu dosya/klasör listesi). iPhone 6,9" alanı için 1290×2796, iPad 13" alanı için 2048×2732.
- **Gizlilik politikası URL'si:** GitHub Pages açıldığında `https://pilotmemorable.github.io/alldrop/privacy.html` (`docs/privacy.html`).
- **Uygulama Gizliliği:** *Veri Toplanmıyor* (Data Not Collected).
- **Kategori:** Yardımcı Programlar (Utilities), ikinci kategori: Verimlilik.
- **Yaş sınırı:** 4+.
- **Şifreleme sorusu:** Uygulama yalnızca standart TLS kullanır; `ITSAppUsesNonExemptEncryption = NO` zaten Info.plist'te var, ek belge gerekmez.

## 4. İncelemeye gönderme

TestFlight'ta derleme göründükten sonra (10–30 dk) sürüm sayfasında derlemeyi seçip **İncelemeye Gönder**'e basın. İnceleme genelde 1–3 gün sürer.

### İncelemede dikkat edilecekler

- **Kural 4.2 (Asgari işlevsellik):** Arayüz web teknolojisiyle çizilse de uygulama kendi aktarım sunucusunu, şifrelemeyi ve dosya yönetimini cihazda çalıştırır. Ret gelirse yanıtta bunu ve yerel ağ aktarımını vurgulayın.
- **İkinci cihaz gerekir:** İnceleme ekibi tek iPhone ile test eder. `notes.txt` bunun için iPhone + tarayıcıdan test adımlarını anlatır.
- **Arka plan:** iOS, uygulama kapalıyken ağ sunucusu çalıştırmaya izin vermez. AllDrop açıkken dosya alır; devam eden aktarım uygulama arka plana geçince kısa bir süre daha sürer.

## Teknik notlar

- En düşük sürüm: iOS/iPadOS 15. Apple Silicon Mac'lerde de çalışır (*Designed for iPad*).
- Alınan dosyalar: Dosyalar uygulaması → iPhone'umda → AllDrop.
- Başka uygulamalardan **Paylaş → AllDrop** ile gelen dosyalar gönderim listesine eklenir.
- iPhone'lar yayın (broadcast) paketi gönderemediğinden uygulama ağı 40 sn'de bir tarar ve bulduğu cihazlara kendini tanıtır.
- Gizlilik manifesti: `AllDrop/PrivacyInfo.xcprivacy` (izleme yok, veri toplama yok).
