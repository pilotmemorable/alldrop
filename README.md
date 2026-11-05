# AllDrop: All Device

**Mac to Windows · Win to Mac · iOS to Android · Android to iOS · All Devices All Transfer**

Windows, Mac, Linux, Android, iPhone ve iPad arasında kablosuz dosya paylaşımı. Kablo yok, internet yok, bulut yok: dosyalar aynı Wi-Fi ağı içinde cihazdan cihaza doğrudan gider.

*English below.*

---

## Özellikler

| | |
|---|---|
| **Platformlar** | Windows 7 · 8 · 8.1 · 10 · 11 (32/64 bit, ARM64), macOS 10.13+ (Intel ve Apple Silicon), Linux, Android 5+, iPhone/iPad (App Store uygulaması ya da tarayıcı) |
| **Güvenlik kodu (OTP)** | Her gönderimde **alıcının ekranında 6 haneli bir kod** çıkar. Gönderen kodu girmeden aktarım başlamaz. 5 yanlış denemede istek iptal olur. Alıcı "Bu cihaza güven" derse o cihazdan bir dahaki sefere kod sorulmaz. Ayarlardan kapatılabilir. |
| **Gizlilik** | Dosyalar hiçbir sunucuya, Google Drive'a ya da internete gitmez. |
| **Şifreleme** | Uygulamalar arası trafik karşılıklı TLS ile şifrelenir. Cihaz kimliği her cihazın kendi anahtarından türetilir. |
| **Kurulum** | Bilgisayarda kurulum yok, tek dosyaya çift tıklanır. Uygulaması olmayan cihazlar QR kodu okutup tarayıcıdan bağlanır. |
| **Hız** | Ağınızın izin verdiği en yüksek hız. Aynı makinedeki testte şifreli bağlantı üzerinden saniyede yaklaşık 300 MB ölçüldü; gerçek sınır Wi-Fi'nizdir. |
| **Ekstralar** | Klasör ve çoklu dosya, sürükle-bırak, Windows'ta sağ tık → Gönder → AllDrop, Android ve iOS'ta **Paylaş** menüsü, ZIP ile toplu indirme, Türkçe/İngilizce arayüz, açık/koyu tema. |

## İndirilecek dosyalar

| Platform | Dosya |
|---|---|
| Windows 64 bit (çoğu bilgisayar) | `AllDrop-Windows-64bit.exe` |
| Windows 32 bit | `AllDrop-Windows-32bit.exe` |
| Windows ARM | `AllDrop-Windows-ARM64.exe` |
| macOS | `AllDrop-macOS.zip` (içinde `AllDrop.app`) |
| Linux | `AllDrop-Linux-x64`, `AllDrop-Linux-ARM64` |
| Android | `AllDrop-Android.apk` (GitHub Actions derler) |
| iPhone / iPad | App Store: **AllDrop: All Device** (`ios/APP-STORE-REHBERI.md`) |

Windows .exe dosyalarının *Özellikler → Ayrıntılar* sekmesinde ürün adı **AllDrop: All Device**, açıklama ve sürüm bilgisi görünür.

## Kullanım

1. AllDrop'u açın. Bilgisayarda tarayıcınızda açılır, telefonda uygulama olarak açılır.
2. Aynı ağdaki AllDrop cihazları kendiliğinden listelenir. Uygulaması olmayan cihazlar ekrandaki QR kodu okutur.
3. Bir cihaza dokunun ve dosyaları seçin (ya da üzerine sürükleyin).
4. Alıcının ekranında görünen **6 haneli kodu** girin. Aktarım başlar.

Alınan dosyalar: bilgisayarda `İndirilenler/AllDrop`, Android'de `Download/AllDrop`, iPhone/iPad'de *Dosyalar → iPhone'umda → AllDrop*.

## İlk açılış (bir kereliğine)

- **Windows:** "Windows bilgisayarınızı korudu" uyarısında *Ek bilgi → Yine de çalıştır*. Güvenlik duvarı sorarsa *İzin ver*.
- **macOS 10.13–14:** sağ tık → *Aç*. **macOS 15+:** *Sistem Ayarları → Gizlilik ve Güvenlik → Yine de Aç*. "Yerel ağ" iznine *İzin Ver* deyin.
- **Android:** APK'yı açarken "bilinmeyen uygulamaları yükleme" izni istenir.

## Proje yapısı

```
engine/           Ortak aktarım motoru (Go): keşif, TLS, güvenlik kodu, web arayüzü (engine/web)
main.go           Windows / macOS / Linux masaüstü programı ve Android motoru
mobile/adkit/     iOS için gomobile API'si
ios/              iPhone/iPad uygulaması (Swift + XcodeGen), mağaza metinleri ve ekran görüntüleri
android/          Android uygulaması (Gradle)
docs/             İndirme sayfası ve gizlilik politikası (GitHub Pages)
.github/workflows Otomatik derleme: release.yml (Windows/Mac/Linux/APK), ios.yml (App Store)
```

## Derleme

```bash
./build.sh                                  # Go 1.20.x: Windows 7+ ve macOS 10.13+ desteği için
cd android && ./gradlew assembleRelease     # Android APK (ya da Android Studio)
./ios/build-engine.sh && cd ios && xcodegen # iOS (Mac + Xcode)
```

GitHub'a gönderip `v1.0.0` gibi bir etiket oluşturduğunuzda tüm dosyalar derlenir ve bir Release yayınlanır. Sırlar tanımlıysa iOS uygulaması App Store Connect'e yüklenir.

## Teknik

- **Keşif:** UDP 45453 yayını, yerel /24 alt ağ taraması, `/peer/hello` ile karşılıklı tanıtım (iPhone'lar yayın gönderemez), elle IP ekleme.
- **Uygulama ↔ uygulama:** TCP 45455 üzerinden karşılıklı TLS. Cihaz kimliği = sertifikanın SHA-256 özeti.
- **Güvenlik kodu:** Kriptografik rastgele 6 hane. Sabit zamanlı karşılaştırılır, en fazla 5 deneme, istek 3 dakikada zaman aşımına uğrar.
- **Tarayıcı misafirleri:** TCP 45454. QR koddaki gizli anahtar olmadan erişim yok. CSRF, Origin ve DNS rebinding korumaları var.
- **Mobil:** Android ve iOS'ta yerel arayüz uygulamaya özel bir belirteçle korunur, telefondaki başka uygulamalar erişemez.

---

## English

**AllDrop: All Device** sends files between Windows, Mac, Linux, Android, iPhone and iPad over your local network, with no cable, internet or cloud.

- **Security code:** each transfer shows a 6-digit code on the receiver. The sender must type it before anything is sent; 5 wrong tries cancel the request. Trusted devices can skip the code.
- App-to-app traffic uses mutually authenticated TLS; a device's ID is the hash of its own certificate.
- Windows 7–11 (x86/x64/ARM64), macOS 10.13+ (universal), Linux, Android 5+, iOS/iPadOS 15+ (App Store). Devices without the app join from the browser via QR code.
- Ports: UDP 45453 (discovery), TCP 45454 (web UI / browser guests), TCP 45455 (encrypted app-to-app).

Third-party code: [go-qrcode](https://github.com/skip2/go-qrcode) (MIT) in `third_party/`.
