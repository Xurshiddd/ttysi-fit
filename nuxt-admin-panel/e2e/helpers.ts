import { expect, type Page } from '@playwright/test'

/// Test admin hisobi. Yaratish:
///   ./run.sh create-admin test@ttyesi.uz "parol123" "Test User"
export const ADMIN = { email: 'test@ttyesi.uz', password: 'parol123' }

/// waitForHydration — Nuxt client-side ishga tushishini kutadi.
///
/// MUHIM: SSR HTML darrov keladi va maydonlar ko'rinadi, lekin Vue hali
/// ulanmagan. O'sha paytda yozilgan matn hidratsiyada v-model boshlang'ich
/// qiymatiga (bo'sh satr) qaytariladi va forma jimgina tozalanadi.
export async function waitForHydration(page: Page) {
  await page.waitForLoadState('networkidle')
  // Nuxt ilova o'rnatilgach `#__nuxt` ichida Vue instansiyasi paydo bo'ladi.
  await page.waitForFunction(() => {
    const el = document.querySelector('#__nuxt') as any
    return !!el && !!el.__vue_app__
  }, { timeout: 20_000 })
}

/// fillStable — hidratsiya matnni yuvib yubormaganini tasdiqlaydi.
/// Qiymat yo'qolsa qayta yozadi (poyga qolgan holatlar uchun himoya).
async function fillStable(page: Page, label: string, value: string) {
  const input = page.getByLabel(label)
  await input.fill(value)
  await expect(input).toHaveValue(value, { timeout: 5_000 }).catch(async () => {
    await input.fill(value)
    await expect(input).toHaveValue(value)
  })
}

/// login — admin panelga haqiqiy forma orqali kiradi.
/// Cookie'ni qo'lda o'rnatmaymiz: login oqimining o'zi ham sinovdan o'tsin.
export async function login(page: Page) {
  await page.goto('/login')
  await waitForHydration(page)

  await fillStable(page, 'Email', ADMIN.email)
  await fillStable(page, 'Parol', ADMIN.password)
  await page.getByRole('button', { name: 'Tizimga kirish' }).click()

  // Login muvaffaqiyatli bo'lsa bosh sahifaga o'tadi.
  await page.waitForURL('/', { timeout: 15_000 })
  await waitForHydration(page)
}

/// expectNoConsoleErrors — sahifada JS xatosi bo'lmasin.
///
/// Buni har bir testda ishlatamiz: Vue'da xato bo'lgan sahifa ko'pincha
/// "ishlayotgandek" ko'rinadi (bo'sh joy yoki eski ma'lumot), faqat konsolda
/// baqiradi. Aynan shunday bug'lar mobil ilovada topilgan edi.
export function collectErrors(page: Page): string[] {
  const errors: string[] = []
  page.on('pageerror', (e) => errors.push(`pageerror: ${e.message}`))
  page.on('console', (m) => {
    if (m.type() === 'error') {
      const t = m.text()
      // Tarmoq 4xx/5xx javoblari alohida tekshiriladi — konsol shovqinini
      // filtrlaymiz (masalan favicon 404).
      if (t.includes('Failed to load resource')) return
      errors.push(`console: ${t}`)
    }
  })
  return errors
}

/// expectClean — xato yo'qligini tekshiradi va bo'lsa hammasini ko'rsatadi.
export function expectClean(errors: string[]) {
  expect(errors, `sahifada JS xatosi:\n${errors.join('\n')}`).toEqual([])
}
