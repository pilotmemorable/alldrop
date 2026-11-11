`rsrc_windows_*.syso` (icon, manifest and version info for the Windows .exe) were generated with
`main.go.txt` built inside a checkout of github.com/akavel/rsrc (it uses that module's internal package):

    git clone https://github.com/akavel/rsrc && mkdir rsrc/cmd/adres && cp main.go.txt rsrc/cmd/adres/main.go
    (cd rsrc && go build -o ../adres ./cmd/adres)
    for a in 386 amd64 arm64; do ./adres -arch $a -manifest packaging/alldrop.manifest -ico packaging/alldrop.ico \
      -o rsrc_windows_$a.syso -version 1.1.2.0 -product "AllDrop: All Device" -desc "AllDrop - All Device Transfer" \
      -comments "Mac to Windows, Win to Mac, iOS to Android, Android to iOS, All Devices All Transfer" \
      -company "AllDrop" -original "AllDrop.exe"; done
