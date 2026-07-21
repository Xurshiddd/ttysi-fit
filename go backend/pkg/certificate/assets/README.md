# Sertifikat aktivlari

## Muhr va imzo (bu papkada YO'Q — ataylab)

Muhr va imzo skani institutning **rasmiy hujjati**, shuning uchun repozitoriyga
qo'shilmaydi: git tarixiga tushgan muhr tasvirini keyin olib tashlab bo'lmaydi
va u soxta hujjat yasashga ishlatilishi mumkin. Fayllar serverda turadi va
`.env` orqali ko'rsatiladi:

```bash
CERT_STAMP_PATH=/srv/ttysi_fit/secrets/muhr.png
CERT_SIGNATURE_PATH=/srv/ttysi_fit/secrets/imzo.png
CERT_SIGNER_NAME="A. A. Aliyev"
CERT_SIGNER_TITLE=Rektor
CERT_SAMPLE_STAMP=false
```

**Talablar:**

| Talab | Sabab |
|-------|-------|
| PNG formati | `fpdf` faqat PNG/JPG/GIF ni o'qiydi |
| **Alfa kanali YO'Q** | shaffof piksellar qora bo'lib chiziladi |
| Kamida ~400×400 px | 24 mm da chop etilganda xira bo'lmasin |

Alfa kanalini olib tashlash:

```bash
convert muhr.png -background white -alpha remove -alpha off muhr-tayyor.png
```

Server **startupda** tekshiradi: fayl yo'q, PNG emas yoki alfa kanali bor bo'lsa
ishga tushmaydi. Buzuq muhr bilan sertifikat bergandan ko'ra shu yaxshi.

### NAMUNA muhri (sinov uchun)

Haqiqiy muhr yo'q bo'lganda `CERT_SAMPLE_STAMP=true` ochiq **"NAMUNA — haqiqiy
muhr emas"** yozuvli aylana chizadi (rasm emas, vektor — fayl kerak emas).
Institutning haqiqiy muhriga **ataylab o'xshatilmagan**: shunda sinov
sertifikati rasmiy hujjat sifatida ishlatib bo'lmaydi.

Production'da bu bayroq yoqilsa server ishga tushmaydi (`config.validate`).

---

## ttysi-logo.png

Institut emblemasi (qalqon). Manba — rasmiy sayt:
`https://ttysi.uz/assets/public/images/logo-pic.svg`

**Nega PNG, SVG emas:** `fpdf` faqat PNG/JPG/GIF ni qo'llab-quvvatlaydi.

**Nega alfa kanali yo'q (fon oq):** `fpdf` bu PNG'ning alfa kanalini o'qimadi —
shaffof piksellar qora bo'lib chizildi va emblema atrofida kulrang to'rtburchak
paydo bo'ldi. Sertifikat foni baribir oq, shuning uchun oq fonga "yassilangan"
rasm ko'rinishda bir xil, lekin ishonchli chiziladi.

### Qayta yaratish

SVG yangilansa (yoki o'lchamni oshirish kerak bo'lsa):

```bash
curl -o logo-pic.svg https://ttysi.uz/assets/public/images/logo-pic.svg
python -c "
import pymupdf
d = pymupdf.open('logo-pic.svg')
# 6x -> 479x566 px. 16 mm kenglikda ~300 DPI (chop etish uchun yetarli).
pix = d[0].get_pixmap(matrix=pymupdf.Matrix(6, 6), alpha=False)
pix.save('ttysi-logo.png')
"
```

`alpha=False` MAJBURIY — yuqoridagi sababga qarang.

## Boshqa mavjud logotiplar

Kerak bo'lsa saytdan olish mumkin:

| Fayl | Tavsif |
|------|--------|
| `logo-pic.svg` | Qalqon emblemasi (shu yerda ishlatilgan) |
| `logo/logo-text_oz.svg` | To'liq nom + "1932-yildan beri" (o'zbekcha) |
| `logo/gerb-transparent.svg` | Davlat gerbi |
| `footer-logo-pic.svg` | Footer varianti |

Nom matni PDF'da **shrift bilan** chiziladi (rasm emas): shunda u aniq chiqadi,
tarjima qilinadi va `certificate.go` dagi konstantadan o'zgartiriladi.
