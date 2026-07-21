import { expect, test, type Page } from '@playwright/test'
import { collectErrors, expectClean, login, waitForHydration } from './helpers'

// FIT Coin do'koni — sovg'alar CRUD va buyurtmalar.
//
// NEGA BU TESTLAR BOR: do'kon PUL harakati bilan ishlaydi (coin yechiladi,
// bekor qilinganda qaytariladi). Formadagi jimgina xato — masalan bo'sh
// "miqdor" maydoni 0 bo'lib ketishi — sovg'ani darrov "tugagan" qilib
// qo'yardi. Shuning uchun to'liq sikl (yaratish → ko'rish → o'chirish)
// haqiqiy backend bilan sinaladi.

async function gotoPage(page: Page, path: string) {
  await page.goto(path)
  await waitForHydration(page)
}

/// expectNoHorizontalScroll — §7.3: sahifa mobilda yon tomonga surilmasin.
async function expectNoHorizontalScroll(page: Page, label: string) {
  const o = await page.evaluate(() => ({
    scroll: document.documentElement.scrollWidth,
    client: document.documentElement.clientWidth
  }))
  expect(o.scroll, `${label}: gorizontal toshdi (${o.scroll} > ${o.client})`)
    .toBeLessThanOrEqual(o.client + 1)
}

test.describe('Sovg‘alar sahifasi', () => {
  test('ro‘yxat yuklanadi va menyuda ko‘rinadi', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)

    await expect(page.getByRole('link', { name: 'Sovg‘alar' })).toBeVisible()
    await gotoPage(page, '/rewards')
    await expect(page.getByRole('heading', { name: 'Sovg‘alar' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Yangi sovg‘a' })).toBeVisible()

    expectClean(errors)
  })

  test('KATEGORIYA backenddan keladi (kodda ro‘yxat yo‘q)', async ({ page }) => {
    await login(page)
    await gotoPage(page, '/rewards')

    await page.getByRole('button', { name: 'Yangi sovg‘a' }).click()
    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()

    // /reward-categories qaytargan variantlar formada bo'lishi kerak.
    const options = dialog.locator('select').first().locator('option')
    await expect(async () => {
      expect(await options.count()).toBeGreaterThanOrEqual(4)
    }).toPass({ timeout: 10_000 })
  })

  test('TO‘LIQ SIKL: yaratish → ro‘yxatda ko‘rinish → o‘chirish', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)
    await gotoPage(page, '/rewards')

    const title = `E2E sovg‘a ${Date.now()}`

    await page.getByRole('button', { name: 'Yangi sovg‘a' }).click()
    const dialog = page.getByRole('dialog')
    await dialog.getByPlaceholder('Masalan: TTYSI_FIT futbolka').fill(title)
    await dialog.locator('input[type="number"]').first().fill('123')
    await dialog.getByRole('button', { name: 'Saqlash' }).click()

    await expect(dialog).toBeHidden({ timeout: 10_000 })
    await expect(page.getByText(title)).toBeVisible({ timeout: 10_000 })
    // Narx to'g'ri saqlanganini tasdiqlaymiz.
    await expect(page.getByRole('row', { name: new RegExp(title) })).toContainText('123')

    // Tozalash — test ma'lumoti DB da qolmasin.
    page.once('dialog', (d) => d.accept())
    await page.getByRole('row', { name: new RegExp(title) })
      .getByRole('button', { name: 'O‘chirish' }).click()
    await expect(page.getByText(title)).toBeHidden({ timeout: 10_000 })

    expectClean(errors)
  })

  test('bo‘sh miqdor CHEKSIZ degani (0 emas)', async ({ page }) => {
    await login(page)
    await gotoPage(page, '/rewards')

    const title = `E2E cheksiz ${Date.now()}`

    await page.getByRole('button', { name: 'Yangi sovg‘a' }).click()
    const dialog = page.getByRole('dialog')
    await dialog.getByPlaceholder('Masalan: TTYSI_FIT futbolka').fill(title)
    await dialog.locator('input[type="number"]').first().fill('10')
    // "Miqdor" maydoniga tegmaymiz — bo'sh qoladi.
    await dialog.getByRole('button', { name: 'Saqlash' }).click()
    await expect(dialog).toBeHidden({ timeout: 10_000 })

    // Bo'sh miqdor 0 ga aylanib "Tugagan" bo'lib qolmasligi kerak.
    const row = page.getByRole('row', { name: new RegExp(title) })
    await expect(row).toBeVisible({ timeout: 10_000 })
    await expect(row).toContainText('∞')
    await expect(row).not.toContainText('Tugagan')

    page.once('dialog', (d) => d.accept())
    await row.getByRole('button', { name: 'O‘chirish' }).click()
  })
})

test.describe('Buyurtmalar sahifasi', () => {
  test('ro‘yxat yuklanadi, default filtr — kutilayotganlar', async ({ page }) => {
    const errors = collectErrors(page)
    await login(page)

    await gotoPage(page, '/redemptions')
    await expect(page.getByRole('heading', { name: 'Buyurtmalar' })).toBeVisible()
    // Adminni birinchi navbatda kutayotgan buyurtmalar qiziqtiradi.
    await expect(page.locator('select')).toHaveValue('pending')

    expectClean(errors)
  })

  test('jadval ustunlari ko‘rinadi', async ({ page }) => {
    await login(page)
    await gotoPage(page, '/redemptions')

    for (const col of ['Kod', 'Sovg‘a', 'Foydalanuvchi', 'Holat']) {
      await expect(page.getByRole('columnheader', { name: col })).toBeVisible()
    }
  })
})

test.describe('Do‘kon — responsivlik (§7.3)', () => {
  for (const w of [375, 768, 1280]) {
    test(`sovg‘alar ${w}px da toshmaydi`, async ({ page }) => {
      await page.setViewportSize({ width: w, height: 900 })
      await login(page)
      await gotoPage(page, '/rewards')

      await expect(page.getByRole('heading', { name: 'Sovg‘alar' })).toBeVisible()
      await expectNoHorizontalScroll(page, `sovg'alar ${w}px`)
    })
  }

  test('buyurtmalar 375px da toshmaydi', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 900 })
    await login(page)
    await gotoPage(page, '/redemptions')

    await expect(page.getByRole('heading', { name: 'Buyurtmalar' })).toBeVisible()
    await expectNoHorizontalScroll(page, 'buyurtmalar 375px')
  })
})
