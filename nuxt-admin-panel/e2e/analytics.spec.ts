import { expect, test, type Page } from '@playwright/test'
import { collectErrors, expectClean, login, waitForHydration } from './helpers'

// Analitika (dashboard grafiklari) va hisobot eksporti testlari.
//
// NEGA BU TESTLAR BOR:
//  1. Grafiklar qo'lda yozilgan SVG/CSS — kutubxona kafolati yo'q. Bo'sh
//     ma'lumot, bitta nuqta yoki nol qiymat matematikani buzishi mumkin
//     (0 ga bo'linish → NaN → SVG `d` atributi yaroqsiz → sahifa oq qoladi).
//  2. Responsivlik CLAUDE.md §7.3 bo'yicha MAJBURIY. Faqat desktopda
//     ko'rib "tayyor" deyish taqiqlangan, shuning uchun uchala kenglik
//     ham shu yerda tekshiriladi.

/// expectNoHorizontalScroll — sahifa gorizontal siljimasin (§7.3).
/// Keng jadval/grafik o'z konteyneridan toshsa telefonda butun sahifa
/// yon tomonga suriladi — bu eng ko'p uchraydigan mobil layout xatosi.
async function expectNoHorizontalScroll(page: Page, label: string) {
  const overflow = await page.evaluate(() => ({
    scroll: document.documentElement.scrollWidth,
    client: document.documentElement.clientWidth
  }))
  expect(
    overflow.scroll,
    `${label}: sahifa gorizontal toshdi (${overflow.scroll}px > ${overflow.client}px)`
  ).toBeLessThanOrEqual(overflow.client + 1)
}

test.describe('Analitika — dashboard', () => {
  test('grafiklar xatosiz chiziladi va raqamlar ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)

    // Umumiy raqamlar bloki
    await expect(page.getByText('Faollik tahlili')).toBeVisible()
    await expect(page.getByText('Jami qadam')).toBeVisible()
    await expect(page.getByText('Kunlik dinamika')).toBeVisible()
    await expect(page.getByText('Fakultetlar kesimi')).toBeVisible()

    // Chiziq grafigi haqiqatan chizilgan bo'lsin.
    const chart = page.getByRole('img', { name: 'Kunlik qadam dinamikasi' })
    await expect(chart).toBeVisible()

    expectClean(errors)
  })

  test('SVG yo‘li yaroqli — NaN yoki Infinity bo‘lmasin', async ({ page }) => {
    await login(page)

    // Chiziq faqat analitika so'rovi qaytgach chiziladi (`v-if="!loading"`).
    // Kutmasdan tekshirilsa test o'zi poyga bo'lib qolardi.
    const path = page.locator('svg[role="img"] path').first()
    await expect(path).toBeAttached({ timeout: 15_000 })

    // Nol qiymatli kunlarda (0 / max) hisob NaN bergan bo'lsa `d` atributi
    // buziladi va brauzer chiziqni umuman chizmaydi — jim xato.
    const paths = await page.locator('svg[role="img"] path').evaluateAll(
      (els) => els.map((e) => e.getAttribute('d') || '')
    )
    expect(paths.length, 'grafik yo‘llari topilmadi').toBeGreaterThan(0)
    for (const d of paths) {
      expect(d, `yaroqsiz SVG yo‘li: ${d}`).not.toMatch(/NaN|Infinity|undefined/)
    }
  })

  test('davrni almashtirish grafikni yangilaydi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)

    await page.getByRole('button', { name: 'Oy', exact: true }).click()
    await page.waitForLoadState('networkidle')

    // Oy davrida kun soni ko'proq — sezgir zonalar soni ham ortadi.
    const days = page.locator('svg[role="img"]').locator('..').locator('button')
    await expect(async () => {
      expect(await days.count()).toBeGreaterThan(7)
    }).toPass({ timeout: 10_000 })

    expectClean(errors)
  })

  test('kunni tanlaganda qiymat ko‘rinadi', async ({ page }) => {
    await login(page)

    // Sezgir zonalar ham ma'lumot kelgach paydo bo'ladi.
    const zones = page.locator('svg[role="img"]').locator('..').locator('button')
    await expect(zones.first()).toBeAttached({ timeout: 15_000 })
    await zones.first().click()

    await expect(page.getByText(/qadam ·/)).toBeVisible()
  })
})

test.describe('Analitika — responsivlik (§7.3)', () => {
  const widths = [
    { w: 375, h: 812, name: 'telefon' },
    { w: 768, h: 1024, name: 'planshet' },
    { w: 1280, h: 900, name: 'desktop' }
  ]

  for (const { w, h, name } of widths) {
    test(`dashboard ${name} (${w}px) da to‘g‘ri ko‘rinadi`, async ({ page }) => {
      const errors = collectErrors(page)
      await page.setViewportSize({ width: w, height: h })
      await login(page)

      await expect(page.getByText('Faollik tahlili')).toBeVisible()
      await expect(page.getByRole('img', { name: 'Kunlik qadam dinamikasi' })).toBeVisible()
      await expectNoHorizontalScroll(page, `dashboard ${w}px`)

      expectClean(errors)
    })
  }

  test(`hisobot sahifasi 375px da to‘g‘ri ko‘rinadi`, async ({ page }) => {
    const errors = collectErrors(page)
    await page.setViewportSize({ width: 375, height: 812 })
    await login(page)

    await page.goto('/reports')
    await waitForHydration(page)

    await expect(page.getByRole('heading', { name: 'Hisobotlar' })).toBeVisible()
    await expect(page.getByRole('button', { name: /CSV yuklab olish/ })).toBeVisible()
    await expectNoHorizontalScroll(page, 'hisobot 375px')

    expectClean(errors)
  })
})

test.describe('Hisobot eksporti', () => {
  test('CSV yuklab olinadi va Excel uchun to‘g‘ri formatda', async ({ page }) => {
    await login(page)
    await page.goto('/reports')
    await waitForHydration(page)

    const downloadPromise = page.waitForEvent('download', { timeout: 30_000 })
    await page.getByRole('button', { name: /CSV yuklab olish/ }).click()
    const download = await downloadPromise

    expect(download.suggestedFilename()).toMatch(/\.csv$/)

    const path = await download.path()
    expect(path, 'fayl saqlanmadi').toBeTruthy()

    const fs = await import('node:fs')
    const buf = fs.readFileSync(path!)

    // BOM — ansiz Excel o'zbek harflarini krakozyabra qiladi.
    expect(
      buf[0] === 0xef && buf[1] === 0xbb && buf[2] === 0xbf,
      'UTF-8 BOM yo‘q — Excel matnni buzadi'
    ).toBe(true)

    const text = buf.toString('utf8')
    // Ajratgich nuqtali vergul (o'zbek/rus lokalidagi Excel uchun).
    expect(text).toContain('F.I.O.;Email;Rol')
  })
})
