import { expect, test } from '@playwright/test'
import { collectErrors, expectClean, login, waitForHydration } from './helpers'

// E'lon yuborish sahifasi.
//
// NEGA BU TESTLAR BOR: e'lon 9 000+ foydalanuvchiga ketadi va QAYTARIB
// OLIB BO'LMAYDI. Shu sababli forma yuborishdan oldin qabul qiluvchilar
// sonini ko'rsatishi va tasdiq so'rashi SHART — bu ikkisi jimgina
// yo'qolib qolsa admin bexosdan butun institutga xabar yuborardi.

test.describe('E‘lonlar sahifasi', () => {
  test('menyuda bor va sahifa ochiladi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)

    await expect(page.getByRole('link', { name: 'E‘lonlar' })).toBeVisible()
    await page.goto('/announcements')
    await waitForHydration(page)

    await expect(page.getByRole('heading', { name: 'E‘lonlar' })).toBeVisible()
    expectClean(errors)
  })

  test('qabul qiluvchilar soni YUBORISHDAN OLDIN ko‘rinadi', async ({ page }) => {
    await login(page)
    await page.goto('/announcements')
    await waitForHydration(page)

    await expect(page.getByText(/Qabul qiluvchilar: /)).toBeVisible({ timeout: 10_000 })
  })

  test('sarlavhasiz yuborib bo‘lmaydi', async ({ page }) => {
    await login(page)
    await page.goto('/announcements')
    await waitForHydration(page)

    // Bo'sh sarlavhada tugma o'chiq bo'lishi kerak.
    await expect(page.getByRole('button', { name: 'Yuborish' })).toBeDisabled()
  })

  test('tasdiqlash oynasi chiqadi va BEKOR qilinsa yuborilmaydi', async ({ page }) => {
    await login(page)
    await page.goto('/announcements')
    await waitForHydration(page)

    await page.getByPlaceholder('Masalan: Universitet krossi 15-avgustda')
      .fill(`E2E sinov ${Date.now()}`)

    // confirm() ni RAD etamiz — hech qanday so'rov ketmasligi kerak.
    let posted = false
    page.on('request', (r) => {
      if (r.method() === 'POST' && r.url().includes('/admin/notifications')) posted = true
    })
    page.once('dialog', (d) => d.dismiss())

    await page.getByRole('button', { name: 'Yuborish' }).click()
    await page.waitForTimeout(1000)

    expect(posted, 'bekor qilinganda ham e‘lon yuborilibdi').toBe(false)
  })

  test('soxta qo‘ng‘iroq tugmasi olib tashlangan', async ({ page }) => {
    await login(page)

    // Ishlamaydigan UI foydalanuvchini chalg'itadi: qo'ng'iroq bosilsa
    // hech narsa qilmaydi va yashil nuqta doim yonib turardi.
    const header = page.locator('header')
    await expect(header.locator('button[title*="ildirishnoma"]')).toHaveCount(0)
  })
})
